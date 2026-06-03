package core

import (
	"errors"
	"fmt"
	"math"

	"github.com/nikolalohinski/gonja/v2/exec"
	"github.com/nikolalohinski/gonja/v2/nodes"
	"github.com/nikolalohinski/gonja/v2/parser"
	"github.com/nikolalohinski/gonja/v2/tokens"
)

var (
	errJinja2goBreak    = errors.New("jinja2go break")
	errJinja2goContinue = errors.New("jinja2go continue")
)

type jinja2goDoControl struct {
	expression nodes.Expression
}

func (control *jinja2goDoControl) Position() *tokens.Token {
	return control.expression.Position()
}

func (control *jinja2goDoControl) String() string {
	token := control.Position()
	return fmt.Sprintf("DoControlStructure(Line=%d Col=%d)", token.Line, token.Col)
}

func (control *jinja2goDoControl) Execute(renderer *exec.Renderer, _ *nodes.ControlStructureBlock) error {
	value := renderer.Eval(control.expression)
	if value.IsError() {
		return value
	}
	return nil
}

func jinja2goDoParser(_ *parser.Parser, args *parser.Parser) (nodes.ControlStructure, error) {
	expression, err := args.ParseExpression()
	if err != nil {
		return nil, err
	}
	if !args.End() {
		return nil, args.Error("Malformed do-tag args.", nil)
	}
	return &jinja2goDoControl{expression: expression}, nil
}

type jinja2goLoopControl struct {
	token *tokens.Token
	err   error
	name  string
}

func (control *jinja2goLoopControl) Position() *tokens.Token {
	return control.token
}

func (control *jinja2goLoopControl) String() string {
	return fmt.Sprintf("%sControlStructure(Line=%d Col=%d)", control.name, control.token.Line, control.token.Col)
}

func (control *jinja2goLoopControl) Execute(_ *exec.Renderer, _ *nodes.ControlStructureBlock) error {
	return control.err
}

func jinja2goBreakParser(_ *parser.Parser, args *parser.Parser) (nodes.ControlStructure, error) {
	if !args.End() {
		return nil, args.Error("Arguments not allowed for break.", nil)
	}
	return &jinja2goLoopControl{token: args.Current(), err: errJinja2goBreak, name: "Break"}, nil
}

func jinja2goContinueParser(_ *parser.Parser, args *parser.Parser) (nodes.ControlStructure, error) {
	if !args.End() {
		return nil, args.Error("Arguments not allowed for continue.", nil)
	}
	return &jinja2goLoopControl{token: args.Current(), err: errJinja2goContinue, name: "Continue"}, nil
}

type jinja2goForControl struct {
	key             string
	value           string
	objectEvaluator nodes.Expression
	ifCondition     nodes.Expression

	bodyWrapper  *nodes.Wrapper
	emptyWrapper *nodes.Wrapper
}

func (control *jinja2goForControl) Position() *tokens.Token {
	return control.bodyWrapper.Position()
}

func (control *jinja2goForControl) String() string {
	token := control.Position()
	return fmt.Sprintf("ForControlStructure(Line=%d Col=%d)", token.Line, token.Col)
}

type jinja2goLoopInfo struct {
	index     int
	index0    int
	length    int
	revindex  int
	revindex0 int
	first     bool
	last      bool
	PrevItem  *exec.Value
	NextItem  *exec.Value
	lastValue *exec.Value
}

func (info *jinja2goLoopInfo) Cycle(args *exec.VarArgs) *exec.Value {
	return args.Args[int(math.Mod(float64(info.index0), float64(len(args.Args))))]
}

func (info *jinja2goLoopInfo) Changed(value *exec.Value) bool {
	same := info.lastValue != nil && value.EqualValueTo(info.lastValue)
	info.lastValue = value
	return !same
}

func (control *jinja2goForControl) Execute(renderer *exec.Renderer, _ *nodes.ControlStructureBlock) error {
	object := renderer.Eval(control.objectEvaluator)
	if object.IsError() {
		return object
	}

	items := exec.NewDict()
	object.Iterate(func(_ int, _ int, key *exec.Value, value *exec.Value) bool {
		sub := renderer.Inherit()
		ctx := sub.Environment.Context
		pair := &exec.Pair{}

		if control.value != "" && !key.IsString() && key.Len() == 2 {
			key.Iterate(func(idx int, _ int, key *exec.Value, _ *exec.Value) bool {
				switch idx {
				case 0:
					ctx.Set(control.key, key)
					pair.Key = key
				case 1:
					ctx.Set(control.value, key)
					pair.Value = key
				}
				return true
			}, func() {})
		} else {
			ctx.Set(control.key, key)
			pair.Key = key
			if value != nil {
				ctx.Set(control.value, value)
				pair.Value = value
			}
		}

		if control.ifCondition != nil && !sub.Eval(control.ifCondition).IsTrue() {
			return true
		}
		items.Pairs = append(items.Pairs, pair)
		return true
	}, func() {})

	length := len(items.Pairs)
	loop := &jinja2goLoopInfo{
		first:  true,
		index0: -1,
		length: length,
	}
	if len(items.Pairs) == 0 && control.emptyWrapper != nil {
		if err := renderer.Inherit().ExecuteWrapper(control.emptyWrapper); err != nil {
			if errors.Is(err, errJinja2goBreak) || errors.Is(err, errJinja2goContinue) {
				return nil
			}
			return err
		}
	}
	for idx, pair := range items.Pairs {
		sub := renderer.Inherit()
		ctx := sub.Environment.Context

		ctx.Set(control.key, pair.Key)
		if pair.Value != nil {
			ctx.Set(control.value, pair.Value)
		}

		ctx.Set("loop", loop)
		loop.index0 = idx
		loop.index = loop.index0 + 1
		if idx == 1 {
			loop.first = false
		}
		if idx+1 == length {
			loop.last = true
		}
		loop.revindex = length - idx
		loop.revindex0 = length - (idx + 1)

		if idx == 0 {
			loop.PrevItem = exec.AsValue(nil)
		} else {
			previous := items.Pairs[idx-1]
			if previous.Value != nil {
				loop.PrevItem = exec.AsValue([2]*exec.Value{previous.Key, previous.Value})
			} else {
				loop.PrevItem = previous.Key
			}
		}

		if idx == length-1 {
			loop.NextItem = exec.AsValue(nil)
		} else {
			next := items.Pairs[idx+1]
			if next.Value != nil {
				loop.NextItem = exec.AsValue([2]*exec.Value{next.Key, next.Value})
			} else {
				loop.NextItem = next.Key
			}
		}

		if err := sub.ExecuteWrapper(control.bodyWrapper); err != nil {
			switch {
			case errors.Is(err, errJinja2goBreak):
				return nil
			case errors.Is(err, errJinja2goContinue):
				continue
			default:
				return err
			}
		}
	}

	return nil
}

func jinja2goForParser(p *parser.Parser, args *parser.Parser) (nodes.ControlStructure, error) {
	control := &jinja2goForControl{}

	var valueToken *tokens.Token
	keyToken := args.Match(tokens.Name)
	if keyToken == nil {
		return nil, args.Error("Expected a key identifier as first argument for 'for'-tag", nil)
	}

	if args.Match(tokens.Comma) != nil {
		valueToken = args.Match(tokens.Name)
		if valueToken == nil {
			return nil, args.Error("Value name must be an identifier.", nil)
		}
	}

	if args.Match(tokens.In) == nil {
		return nil, args.Error("Expected keyword 'in'.", nil)
	}

	objectEvaluator, err := args.ParseExpression()
	if err != nil {
		return nil, err
	}
	control.objectEvaluator = objectEvaluator
	control.key = keyToken.Val
	if valueToken != nil {
		control.value = valueToken.Val
	}

	if args.MatchName("if") != nil {
		ifCondition, err := args.ParseExpression()
		if err != nil {
			return nil, err
		}
		control.ifCondition = ifCondition
	}

	if !args.End() {
		return nil, args.Error("Malformed for-loop args.", nil)
	}

	wrapper, endargs, err := p.WrapUntil("else", "endfor")
	if err != nil {
		return nil, err
	}
	control.bodyWrapper = wrapper

	if !endargs.End() {
		return nil, endargs.Error("Arguments not allowed here.", nil)
	}

	if wrapper.EndTag == "else" {
		wrapper, endargs, err = p.WrapUntil("endfor")
		if err != nil {
			return nil, err
		}
		control.emptyWrapper = wrapper

		if !endargs.End() {
			return nil, endargs.Error("Arguments not allowed here.", nil)
		}
	}

	return control, nil
}
