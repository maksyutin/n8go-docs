(function () {
    'use strict';

    var SCHEME_KEY = 'md-color-scheme';
    var SCHEME_DARK = 'slate';
    var SCHEME_LIGHT = 'default';

    // ---------- Color scheme (dark/light) ----------

    function getScheme() {
        return localStorage.getItem(SCHEME_KEY) || SCHEME_LIGHT;
    }

    function applyScheme(scheme) {
        document.documentElement.setAttribute('data-md-color-scheme', scheme);
        localStorage.setItem(SCHEME_KEY, scheme);
    }

    applyScheme(getScheme());

    function getMermaidTheme() {
        var scheme = document.documentElement.getAttribute('data-md-color-scheme');
        return scheme === SCHEME_DARK ? 'dark' : 'default';
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
        // Reset already-rendered diagrams back to source so mermaid can re-render
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

    document.addEventListener('DOMContentLoaded', function () {

        // --- Palette toggle button ---
        var paletteBtn = document.getElementById('__palette');
        if (paletteBtn) {
            paletteBtn.addEventListener('click', function () {
                var current = document.documentElement.getAttribute('data-md-color-scheme');
                applyScheme(current === SCHEME_DARK ? SCHEME_LIGHT : SCHEME_DARK);
                renderMermaid();
            });
        }

        // --- Mobile drawer: close on overlay click ---
        var drawerToggle = document.getElementById('__drawer');
        var overlay = document.querySelector('.md-overlay');
        if (overlay && drawerToggle) {
            overlay.addEventListener('click', function () {
                drawerToggle.checked = false;
            });
        }

        // --- TOC: highlight active section on scroll ---
        function initTocHighlight() {
            var tocLinks = document.querySelectorAll('.md-nav--secondary .md-nav__link[data-toc-id]');
            if (tocLinks.length === 0) return;

            var headings = [];
            tocLinks.forEach(function (link) {
                var id = link.getAttribute('data-toc-id');
                var el = document.getElementById(id);
                if (el) headings.push({ el: el, link: link });
            });

            function onScroll() {
                var scrollY = window.scrollY;
                var active = null;
                for (var i = headings.length - 1; i >= 0; i--) {
                    var top = headings[i].el.getBoundingClientRect().top + scrollY;
                    if (scrollY >= top - 80) {
                        active = headings[i];
                        break;
                    }
                }
                tocLinks.forEach(function (l) { l.classList.remove('md-nav__link--toc-active'); });
                if (active) active.link.classList.add('md-nav__link--toc-active');
            }

            // Remove previous scroll listener by replacing it
            window._tocScrollHandler && window.removeEventListener('scroll', window._tocScrollHandler);
            window._tocScrollHandler = onScroll;
            window.addEventListener('scroll', onScroll, { passive: true });
            onScroll();
        }

        initTocHighlight();
        initMermaid();
        initMermaidLightbox();

        // --- SPA-style navigation ---
        var mainContent = document.getElementById('main-content');

        // Mark active tab and sidebar section based on current href
        function setActiveTab(href) {
            var tabItems = document.querySelectorAll('.md-tabs__item');
            tabItems.forEach(function (item) {
                var link = item.querySelector('.md-tabs__link');
                if (!link) return;
                var linkHref = link.getAttribute('href');
                // Match if current path starts with same segment as tab link
                var tabSegment = linkHref ? linkHref.replace(/^\//, '').split('/')[0] : '';
                var curSegment = href.replace(/^\//, '').split('/')[0];
                if (tabSegment && tabSegment === curSegment) {
                    item.classList.add('md-tabs__item--active');
                } else {
                    item.classList.remove('md-tabs__item--active');
                }
            });
        }

        function setActiveNavLink(href) {
            var navLinks = document.querySelectorAll('.md-nav--primary .md-nav__link[href]');
            navLinks.forEach(function (l) {
                l.classList.remove('md-nav__link--active');
                var linkHref = l.getAttribute('href');
                if (linkHref && href.endsWith(linkHref.replace(/^\//, ''))) {
                    l.classList.add('md-nav__link--active');
                    // Expand parent nested section
                    var item = l.closest('.md-nav__item--nested');
                    if (item) {
                        var toggle = item.querySelector('.md-nav__toggle');
                        if (toggle) toggle.checked = true;
                    }
                }
            });
        }

        if (mainContent) {
            document.addEventListener('click', function (e) {
                var link = e.target.closest('a.md-nav__link, a.md-tabs__link');
                if (!link) return;
                var href = link.getAttribute('href');
                if (!href || href.startsWith('http') || href.startsWith('#') || href.startsWith('mailto:')) return;

                e.preventDefault();
                var url = href.startsWith('/') ? href : ('/' + href);

                window.history.pushState(null, null, url);
                navigateTo(url, href);
            });
        }

        function navigateTo(url, href) {
            fetch(url)
                .then(function (r) { return r.text(); })
                .then(function (html) {
                    var parser = new DOMParser();
                    var doc = parser.parseFromString(html, 'text/html');

                    document.title = doc.title;

                    // Update main content
                    var newContent = doc.getElementById('main-content');
                    var mainEl = document.getElementById('main-content');
                    if (newContent && mainEl) {
                        mainEl.innerHTML = newContent.innerHTML;
                    }

                    // Update TOC
                    var newToc = doc.getElementById('md-toc-list');
                    var tocEl = document.getElementById('md-toc-list');
                    if (tocEl) {
                        tocEl.innerHTML = newToc ? newToc.innerHTML : '';
                    }

                    // Update left sidebar (primary nav — children of active tab)
                    var newNav = doc.getElementById('md-nav-primary');
                    var navEl = document.getElementById('md-nav-primary');
                    if (navEl && newNav) {
                        navEl.innerHTML = newNav.innerHTML;
                    } else if (!navEl && newNav) {
                        // sidebar inner might need full replacement
                        var sidebarInner = document.querySelector('.md-sidebar--primary .md-sidebar__inner');
                        if (sidebarInner) {
                            var newSidebarInner = doc.querySelector('.md-sidebar--primary .md-sidebar__inner');
                            if (newSidebarInner) sidebarInner.innerHTML = newSidebarInner.innerHTML;
                        }
                    }

                    // Update tabs active state
                    var newTabs = doc.getElementById('md-tabs');
                    var tabsEl = document.getElementById('md-tabs');
                    if (tabsEl && newTabs) {
                        tabsEl.innerHTML = newTabs.innerHTML;
                    }

                    setActiveTab(href);
                    setActiveNavLink(href);
                    window.scrollTo(0, 0);
                    initTocHighlight();
                    initMermaid();

                    // Close mobile drawer
                    var drawerChk = document.getElementById('__drawer');
                    if (drawerChk) drawerChk.checked = false;
                })
                .catch(function () {
                    window.location.href = url;
                });
        }

        // --- Handle browser back/forward ---
        window.addEventListener('popstate', function () {
            window.location.reload();
        });

    });
})();
