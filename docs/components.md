# Components

Built-in UI components available in the **default** theme. Components are rendered as custom HTML elements and activated by theme JavaScript.

---

## Alert

Displays an inline notification box with contextual styling.

### Usage

```html
<n8go-alert type="info" message="This is an info alert"></n8go-alert>
```

### Props

| Prop | Values | Description |
|------|--------|-------------|
| `type` | <span class="badge bg-info text-dark">info</span> <span class="badge bg-success">success</span> <span class="badge bg-warning text-dark">warn</span> <span class="badge bg-danger">error</span> | Controls the Bootstrap alert variant, title, and icon. |
| `message` | string | The text to display inside the alert box. |

### Examples

```html
<n8go-alert type="info"    message="Informational message"></n8go-alert>
<n8go-alert type="success" message="Operation completed successfully"></n8go-alert>
<n8go-alert type="warn"    message="Proceed with caution"></n8go-alert>
<n8go-alert type="error"   message="Something went wrong"></n8go-alert>
```

<n8go-alert type="info"    message="Informational message"></n8go-alert>
<n8go-alert type="success" message="Operation completed successfully"></n8go-alert>
<n8go-alert type="warn"    message="Proceed with caution"></n8go-alert>
<n8go-alert type="error"   message="Something went wrong"></n8go-alert>

---

## Theme compatibility

Components in this section are implemented in the **default** theme only. The **material** theme does not bundle these custom elements. If you are building a custom theme, include `themes/default/js/theme.js` or re-implement the components in your own JavaScript.
