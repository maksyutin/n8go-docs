(function () {
    'use strict';

    var ROOT_PATH = (function () {
        var scripts = document.querySelectorAll('script[src*="search.js"]');
        if (scripts.length) {
            var src = scripts[scripts.length - 1].getAttribute('src');
            return src.replace(/js\/search\.js$/, '');
        }
        return '';
    })();

    var searchInput = document.getElementById('md-search-input');
    var resultList = document.getElementById('md-search-result-list');
    var resultMeta = document.querySelector('.md-search-result__meta');
    var searchCheckbox = document.getElementById('__search');

    if (!searchInput || !resultList) return;

    var index = null;
    var indexData = [];

    // Load FlexSearch index
    function loadIndex() {
        if (index !== null) return Promise.resolve();
        return fetch(ROOT_PATH + 'search/index.json')
            .then(function (r) { return r.json(); })
            .then(function (data) {
                indexData = data;
                if (typeof FlexSearch !== 'undefined') {
                    index = new FlexSearch.Document({
                        tokenize: 'forward',
                        cache: true,
                        document: {
                            id: 'Url',
                            index: ['Title', 'Content'],
                            store: ['Title', 'Content', 'Url'],
                        }
                    });
                    data.forEach(function (item) { index.add(item); });
                } else {
                    // Fallback: simple substring search
                    index = { type: 'simple' };
                }
            })
            .catch(function () {
                index = { type: 'simple' };
            });
    }

    function simpleSearch(query) {
        var q = query.toLowerCase();
        return indexData.filter(function (item) {
            return (item.Title && item.Title.toLowerCase().includes(q)) ||
                   (item.Content && item.Content.toLowerCase().includes(q));
        }).slice(0, 10);
    }

    function highlight(text, query) {
        if (!text || !query) return escHtml(text || '');
        var esc = query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
        return escHtml(text).replace(new RegExp('(' + esc + ')', 'gi'), '<mark>$1</mark>');
    }

    function escHtml(str) {
        return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    }

    function truncate(text, maxLen) {
        if (!text) return '';
        if (text.length <= maxLen) return text;
        return text.substring(0, maxLen) + '…';
    }

    function renderResults(results, query) {
        if (results.length === 0) {
            resultMeta.textContent = 'No results for "' + query + '"';
            resultList.innerHTML = '';
            return;
        }
        resultMeta.textContent = results.length + ' result' + (results.length === 1 ? '' : 's') + ' for "' + query + '"';
        var html = '';
        results.forEach(function (item) {
            var title = item.Title || 'Untitled';
            var content = truncate(item.Content || '', 120);
            html += '<li class="md-search-result__item">' +
                '<a href="' + escHtml(item.Url) + '">' +
                '<p class="md-search-result__title">' + highlight(title, query) + '</p>' +
                '<p class="md-search-result__teaser">' + highlight(content, query) + '</p>' +
                '</a></li>';
        });
        resultList.innerHTML = html;
    }

    function doSearch(query) {
        if (!query || query.length < 2) {
            resultMeta.textContent = 'Type to start searching';
            resultList.innerHTML = '';
            return;
        }
        loadIndex().then(function () {
            var results;
            if (index && index.type !== 'simple' && typeof FlexSearch !== 'undefined') {
                var raw = index.search(query, { enrich: true, limit: 10 });
                var seen = {};
                results = [];
                raw.forEach(function (field) {
                    (field.result || []).forEach(function (r) {
                        var url = r.doc ? r.doc.Url : r.Url;
                        if (!seen[url]) {
                            seen[url] = true;
                            results.push(r.doc || r);
                        }
                    });
                });
            } else {
                results = simpleSearch(query);
            }
            renderResults(results, query);
        });
    }

    var debounceTimer;
    searchInput.addEventListener('input', function () {
        clearTimeout(debounceTimer);
        var q = this.value.trim();
        debounceTimer = setTimeout(function () { doSearch(q); }, 150);
    });

    // Focus input when search opens
    if (searchCheckbox) {
        searchCheckbox.addEventListener('change', function () {
            if (this.checked) {
                setTimeout(function () { searchInput.focus(); }, 50);
                loadIndex();
                if (searchInput.value) doSearch(searchInput.value.trim());
            }
        });
    }

    // Keyboard: Escape closes search
    document.addEventListener('keydown', function (e) {
        if (e.key === 'Escape' && searchCheckbox && searchCheckbox.checked) {
            searchCheckbox.checked = false;
        }
        // Ctrl+K / Cmd+K opens search
        if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
            e.preventDefault();
            if (searchCheckbox) searchCheckbox.checked = true;
        }
    });

    // Clear button
    var clearBtn = document.querySelector('.md-search__form button[type="reset"]');
    if (clearBtn) {
        clearBtn.addEventListener('click', function () {
            searchInput.value = '';
            resultMeta.textContent = 'Type to start searching';
            resultList.innerHTML = '';
            searchInput.focus();
        });
    }
})();
