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
        syncToggleLabel();
    }

    function syncToggleLabel() {
        var toggleBtn = document.getElementById('toggle_btn') || document.getElementById('__palette');
        if (!toggleBtn) return;
        var current = document.documentElement.getAttribute('data-md-color-scheme');
        toggleBtn.setAttribute('aria-label', current === SCHEME_DARK ? 'Switch to light theme' : 'Switch to dark theme');
        toggleBtn.setAttribute('title', current === SCHEME_DARK ? 'Switch to light mode' : 'Switch to dark mode');
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
        document.addEventListener('click', function (e) {
            var paletteBtn = e.target.closest('#toggle_btn, #__palette');
            if (paletteBtn) {
                var current = document.documentElement.getAttribute('data-md-color-scheme');
                applyScheme(current === SCHEME_DARK ? SCHEME_LIGHT : SCHEME_DARK);
                renderMermaid();
            }
        });

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

        function normalizePath(pathname) {
            if (!pathname) return '/';
            return pathname.endsWith('/') ? pathname : pathname + '/';
        }

        function isInternalNavigationLink(link) {
            if (!link || !link.href) return false;
            if (link.hasAttribute('download')) return false;
            if (link.target && link.target !== '_self') return false;
            var href = link.getAttribute('href');
            if (!href || href.startsWith('#') || href.startsWith('mailto:') || href.startsWith('tel:')) return false;
            var url = new URL(link.href, window.location.href);
            return url.origin === window.location.origin;
        }

        function pathMatches(currentPath, candidatePath) {
            if (!candidatePath) return false;
            if (currentPath === candidatePath) return true;
            if (candidatePath === '/') return false;
            return currentPath.indexOf(candidatePath) === 0;
        }

        function findBestPathMatch(links, currentPath) {
            var best = null;
            var bestLength = -1;

            links.forEach(function (link) {
                var linkPath = normalizePath(new URL(link.href, window.location.href).pathname);
                if (!pathMatches(currentPath, linkPath)) return;
                if (linkPath.length > bestLength) {
                    best = link;
                    bestLength = linkPath.length;
                }
            });

            return best;
        }

        // Mark active tab and sidebar section based on current href
        function setActiveTab(pathname) {
            var currentPath = normalizePath(pathname);
            var tabItems = document.querySelectorAll('.md-tabs__item');
            var tabLinks = Array.prototype.slice.call(document.querySelectorAll('.md-tabs__link[href]'));
            var activeLink = findBestPathMatch(tabLinks, currentPath);

            tabItems.forEach(function (item) {
                item.classList.remove('md-tabs__item--active');
            });

            if (activeLink) {
                var activeItem = activeLink.closest('.md-tabs__item');
                if (activeItem) activeItem.classList.add('md-tabs__item--active');
            }
        }

        function setActiveNavLink(pathname) {
            var currentPath = normalizePath(pathname);
            var navLinks = Array.prototype.slice.call(document.querySelectorAll('.md-nav--primary .md-nav__link[href]'));
            navLinks.forEach(function (l) {
                l.classList.remove('md-nav__link--active');
                l.removeAttribute('aria-current');
            });

            var activeLink = findBestPathMatch(navLinks, currentPath);
            if (activeLink) {
                activeLink.classList.add('md-nav__link--active');
                activeLink.setAttribute('aria-current', 'page');

                // Expand all parent nested sections
                var parent = activeLink.closest('.md-nav__item--nested');
                while (parent) {
                    var toggle = parent.querySelector(':scope > .md-nav__toggle');
                    if (toggle) toggle.checked = true;
                    parent = parent.parentElement ? parent.parentElement.closest('.md-nav__item--nested') : null;
                }
            }
        }

        function syncHead(doc) {
            document.title = doc.title;

            var nextCanonical = doc.querySelector('link[rel="canonical"]');
            var currentCanonical = document.querySelector('link[rel="canonical"]');
            if (nextCanonical && currentCanonical) {
                currentCanonical.setAttribute('href', nextCanonical.getAttribute('href'));
            } else if (nextCanonical && !currentCanonical) {
                document.head.appendChild(nextCanonical.cloneNode(true));
            } else if (!nextCanonical && currentCanonical) {
                currentCanonical.remove();
            }
        }

        function replacePrimaryNav(doc) {
            var navEl = document.getElementById('md-nav-primary');
            var newNav = doc.getElementById('md-nav-primary');
            var sidebarInner = document.querySelector('.md-sidebar--primary .md-sidebar__inner');
            var newSidebarInner = doc.querySelector('.md-sidebar--primary .md-sidebar__inner');

            if (navEl && newNav) {
                navEl.innerHTML = newNav.innerHTML;
                return;
            }

            if (sidebarInner && newSidebarInner) {
                sidebarInner.innerHTML = newSidebarInner.innerHTML;
                return;
            }

            if (sidebarInner && !newSidebarInner) {
                sidebarInner.innerHTML = '';
            }
        }

        function replaceToc(doc) {
            var tocSidebar = document.querySelector('.md-sidebar--secondary .md-sidebar__inner');
            var newTocSidebar = doc.querySelector('.md-sidebar--secondary .md-sidebar__inner');

            if (tocSidebar && newTocSidebar) {
                tocSidebar.innerHTML = newTocSidebar.innerHTML;
                return;
            }

            if (tocSidebar && !newTocSidebar) {
                tocSidebar.innerHTML = '';
            }
        }

        function navigateTo(url, options) {
            options = options || {};
            var historyMode = options.historyMode || 'push';

            fetch(url, { headers: { 'X-Requested-With': 'XMLHttpRequest' } })
                .then(function (r) { return r.text(); })
                .then(function (html) {
                    var parser = new DOMParser();
                    var doc = parser.parseFromString(html, 'text/html');
                    var currentUrl = new URL(url, window.location.href);

                    // Update main content
                    var newContent = doc.getElementById('main-content');
                    var mainEl = document.getElementById('main-content');
                    if (newContent && mainEl) {
                        mainEl.innerHTML = newContent.innerHTML;
                    }

                    if (!newContent || !mainEl) {
                        window.location.href = currentUrl.href;
                        return;
                    }

                    if (historyMode === 'push') {
                        window.history.pushState(null, '', currentUrl.href);
                    }
                    if (historyMode === 'replace') {
                        window.history.replaceState(null, '', currentUrl.href);
                    }

                    syncHead(doc);
                    replacePrimaryNav(doc);
                    replaceToc(doc);

                    setActiveTab(currentUrl.pathname);
                    setActiveNavLink(currentUrl.pathname);
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

        if (mainContent) {
            document.addEventListener('click', function (e) {
                if (e.defaultPrevented || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;

                var link = e.target.closest('a[href]');
                if (!isInternalNavigationLink(link)) return;

                var url = new URL(link.href, window.location.href);
                if (url.hash && normalizePath(url.pathname) === normalizePath(window.location.pathname)) {
                    return;
                }

                e.preventDefault();
                navigateTo(url.href, { historyMode: 'push' });
            });
        }

        // --- Handle browser back/forward ---
        window.addEventListener('popstate', function () {
            navigateTo(window.location.href, { historyMode: 'replace' });
        });

    });
})();
