(function () {
    function getMermaidTheme() {
        return document.documentElement.getAttribute('data-theme') === 'dark' ? 'dark' : 'default';
    }

    function initMermaid() {
        if (typeof mermaid === 'undefined') return;
        var blocks = document.querySelectorAll('pre code.language-mermaid');
        blocks.forEach(function (code) {
            var pre = code.parentNode;
            var div = document.createElement('div');
            div.className = 'mermaid';
            div.setAttribute('data-mermaid-src', code.textContent);
            div.textContent = code.textContent;
            pre.parentNode.replaceChild(div, pre);
        });
        renderMermaid();
    }

    function renderMermaid() {
        if (typeof mermaid === 'undefined') return;
        var diagrams = document.querySelectorAll('.mermaid');
        if (!diagrams.length) return;
        diagrams.forEach(function (div) {
            var src = div.getAttribute('data-mermaid-src');
            if (src && div.getAttribute('data-processed')) {
                div.removeAttribute('data-processed');
                div.innerHTML = '';
                div.textContent = src;
            }
        });
        mermaid.initialize({ startOnLoad: false, theme: getMermaidTheme() });
        mermaid.run({ querySelector: '.mermaid' });
    }

    var _lightbox = null;
    function getMermaidLightbox() {
        if (_lightbox) return _lightbox;
        _lightbox = document.createElement('div');
        _lightbox.className = 'mermaid-lightbox';
        _lightbox.innerHTML = '<button class="mermaid-lightbox__close" aria-label="Close">&times;</button><div class="mermaid-lightbox__inner"></div>';
        document.body.appendChild(_lightbox);
        var inner = _lightbox.querySelector('.mermaid-lightbox__inner');
        _lightbox.addEventListener('click', function (e) {
            if (e.target === _lightbox || e.target.closest('.mermaid-lightbox__close')) {
                _lightbox.classList.remove('is-open');
                inner.innerHTML = '';
            }
        });
        document.addEventListener('keydown', function (e) {
            if (e.key === 'Escape') {
                _lightbox.classList.remove('is-open');
                inner.innerHTML = '';
            }
        });
        return _lightbox;
    }

    function initMermaidLightbox() {
        document.addEventListener('click', function (e) {
            var diagram = e.target.closest('.mermaid');
            if (!diagram) return;
            var svg = diagram.querySelector('svg');
            if (!svg) return;
            var lb = getMermaidLightbox();
            var inner = lb.querySelector('.mermaid-lightbox__inner');
            inner.innerHTML = '';
            var clone = svg.cloneNode(true);
            clone.removeAttribute('width');
            clone.removeAttribute('height');
            clone.style.width = '100%';
            clone.style.height = 'auto';
            inner.appendChild(clone);
            lb.classList.add('is-open');
        });
    }

    var nav = document.getElementById('nav');
    var mainContent = document.getElementById('main-content');
    var menuButton = document.getElementById('menu-button');
    menuButton.onclick = function () {
        nav.classList.toggle('open');
        mainContent.classList.toggle('nav-open');
        menuButton.setAttribute('aria-expanded', nav.classList.contains('open') ? 'true' : 'false');
    }

    function setHTML(o, html, clear) {
        if (clear) o.innerHTML = "";
    
        // Generate a parseable object with the html:
        var dv = document.createElement("div");
        dv.innerHTML = html;
    
        // Handle edge case where innerHTML contains no tags, just text:
        if (dv.children.length===0){ o.innerHTML = html; return; }
    
        for (var i = 0; i < dv.children.length; i++) {
            var c = dv.children[i];
    
            // n: new node with the same type as c
            var n = document.createElement(c.nodeName);
    
            // copy all attributes from c to n
            for (var j = 0; j < c.attributes.length; j++)
                n.setAttribute(c.attributes[j].nodeName, c.attributes[j].nodeValue);
    
            // If current node is a leaf, just copy the appropriate property (text or innerHTML)
            if (c.children.length == 0)
            {
                switch (c.nodeName)
                {
                    case "SCRIPT":
                        if (c.text) n.text = c.text;
                        break;
                    default:
                        if (c.innerHTML) n.innerHTML = c.innerHTML;
                        break;
                }
            }
            // If current node has sub nodes, call itself recursively:
            else setHTML(n, c.innerHTML, false);
            o.appendChild(n);
        }
    }

    function navigateTo(pathname, urlForHistory, shouldPushState) {
        fetch(pathname)
            .then(function (response) { return response.text(); })
            .then(function (html) {
                var parser = new DOMParser();
                var doc = parser.parseFromString(html, 'text/html');

                document.title = doc.title;
                var newFavicon = doc.querySelector('link[rel~="icon"]');
                var currentFavicon = document.querySelector('link[rel~="icon"]');
                if (newFavicon && currentFavicon) {
                    currentFavicon.setAttribute('href', newFavicon.getAttribute('href'));
                }

                var newContent = doc.getElementById('main-content');
                if (newContent) mainContent.innerHTML = newContent.innerHTML;

                // Replace entire nav with the fetched page's nav (correct relative links)
                var newNav = doc.getElementById('nav');
                if (newNav) nav.innerHTML = newNav.innerHTML;

                var toc = document.getElementById('toc');
                var newToc = doc.getElementById('toc');
                if (toc && newToc) toc.innerHTML = newToc.innerHTML;

                if (shouldPushState && urlForHistory) {
                    window.history.pushState(null, null, urlForHistory);
                }

                initMermaid();
            })
            .catch(function () {
                window.location.href = urlForHistory || pathname;
            });
    }

    // Event delegation — works even after nav innerHTML is replaced
    document.addEventListener('click', function (e) {
        var link = e.target.closest('#nav a');
        if (!link) return;
        var rawHref = link.getAttribute('href');
        if (!rawHref || rawHref.startsWith('http') || rawHref.startsWith('#') || rawHref.startsWith('mailto:')) return;
        e.preventDefault();
        var resolved = new URL(rawHref, window.location.href);
        navigateTo(resolved.pathname, resolved.href, true);
    });

    window.addEventListener('popstate', function () {
        navigateTo(window.location.pathname, window.location.href, false);
    });

    initMermaid();
    initMermaidLightbox();

    function setTheme(theme) {
        var toggleBtn = document.getElementById('toggle_btn');
        if (theme === 'dark') {
            document.documentElement.setAttribute('data-theme', 'dark');
            if (toggleBtn) {
                toggleBtn.setAttribute('aria-label', 'Switch to light theme');
            }
        } else {
            document.documentElement.removeAttribute('data-theme');
            if (toggleBtn) {
                toggleBtn.setAttribute('aria-label', 'Switch to dark theme');
            }
        }
    }
    if (localStorage.getItem('theme') === 'dark') {
        setTheme('dark');
    } else {
        setTheme('light');
    }

    // Use delegation so the toggle still works after nav HTML is replaced.
    document.addEventListener('click', function (e) {
        var toggleBtn = e.target.closest('#toggle_btn');
        if (!toggleBtn) return;
        if (document.documentElement.getAttribute('data-theme') === 'dark') {
            setTheme('light');
            localStorage.setItem('theme', 'light');
        } else {
            setTheme('dark');
            localStorage.setItem('theme', 'dark');
        }
        renderMermaid();
    });
})();
