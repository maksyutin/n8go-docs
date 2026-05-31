(async function () {
    var ROOT_PATH = (function () {
        var scripts = document.querySelectorAll('script[src*="search.js"]');
        if (scripts.length) {
            var src = scripts[scripts.length - 1].getAttribute('src');
            return src.replace(/js\/search\.js$/, '');
        }
        return '/';
    })();

    async function fetchSearchIndex() {
        const response = await fetch(ROOT_PATH + 'search/index.json');
        if (!response.ok) throw new Error('Search index not found: ' + response.status);
        return response.json();
    }
    var searchIndex = await fetchSearchIndex()

    const fuse = new Fuse(searchIndex, {
        keys: ['Title','Content'],
        minMatchCharLength: 3,
        includeMatches: true,
    });
    console.log('Fuse.js Initialized');

    const appendHighlightedText = (parent, inputText, regions = [], highlightClassName = 'highlight') => {
        const text = String(inputText || '');
        let nextTextIndex = 0;

        regions.forEach((region) => {
            const start = Math.max(region[0], nextTextIndex);
            const end = Math.min(region[1] + 1, text.length);
            if (start > nextTextIndex) {
                parent.appendChild(document.createTextNode(text.substring(nextTextIndex, start)));
            }
            if (end > start) {
                const mark = document.createElement('span');
                mark.className = highlightClassName;
                mark.textContent = text.substring(start, end);
                parent.appendChild(mark);
            }
            nextTextIndex = end;
        });

        if (nextTextIndex < text.length) {
            parent.appendChild(document.createTextNode(text.substring(nextTextIndex)));
        }
    };

    const matchesForKey = (matches = [], key) => {
        const match = matches.find((item) => item.key === key);
        return match ? match.indices : [];
    };

    const resultUrl = (url) => {
        let cleanUrl = String(url || '').replace(/[\u0000-\u001F\u007F]/g, '');
        if (/^[a-zA-Z][a-zA-Z\d+.-]*:/.test(cleanUrl)) {
            cleanUrl = './' + cleanUrl;
        }
        return ROOT_PATH + cleanUrl;
    };

    const appendEmptyResult = (parent) => {
        const wrapper = document.createElement('center');
        const message = document.createElement('span');
        message.style.color = '#7f8497';
        message.style.fontSize = '0.9em';
        message.style.textAlign = 'center';
        message.style.paddingTop = '10px';
        message.style.paddingBottom = '10px';
        message.textContent = 'No results found. Make sure to have atleast 3 characters.';
        wrapper.appendChild(message);
        parent.appendChild(wrapper);
    };

    const appendSearchResult = (parent, result) => {
        const item = result.item || {};
        const li = document.createElement('li');
        const link = document.createElement('a');
        const title = document.createElement('span');
        const content = document.createElement('p');

        link.className = 'search-result';
        link.setAttribute('href', resultUrl(item.Url));
        appendHighlightedText(title, item.Title, matchesForKey(result.matches, 'Title'));
        appendHighlightedText(content, item.Content, matchesForKey(result.matches, 'Content'));

        link.appendChild(title);
        link.appendChild(content);
        li.appendChild(link);
        parent.appendChild(li);
    };

    var searchInput = document.getElementById('search-input');

    searchInput.addEventListener('keyup', function (e) {
        var searchResults = document.getElementById('search-results');
        var searchValue = e.target.value;
        var results = fuse.search(searchValue);
        searchResults.textContent = '';
        if (results.length) {
            results.forEach((result) => appendSearchResult(searchResults, result));
        }
        if (!results.length) {
            appendEmptyResult(searchResults);
        }
    });
    var searchContainer = document.getElementById('search-container');
    var searchBtn = document.getElementById('search_btn');
    searchBtn.onclick = function () {
        searchContainer.style.display = 'block';
        searchBtn.setAttribute('aria-expanded', 'true');
        searchInput.focus();
    }

    var searchModalCloseBtn = document.getElementById('search-modal-close');
    searchModalCloseBtn.onclick = function () {
        searchContainer.style.display = 'none';
        searchBtn.setAttribute('aria-expanded', 'false');
    }
})();
