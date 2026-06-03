class Alert extends HTMLElement {
    constructor() { super(); }
    render() {
        var type = this.resolveType(this.getAttribute('type'));
        var title = this.resolveText(type);
        var bootstrapType = this.resolveBootstrapType(type);
        var message = this.escapeHTML(this.getAttribute('message') || '');

        this.innerHTML = `
            <div class="alert alert-${bootstrapType} n8go-alert" role="alert">
                <div class="d-flex">
                    <div class="flex-shrink-0">
                        ${this.resolveIcon(type)}
                    </div>
                    <div style="margin-left: .75rem;">
                        <h3 class="alert-heading n8go-alert__title">${title}</h3>
                        <div class="n8go-alert__message">
                            ${message}
                        </div>
                    </div>
                </div>
            </div>
        `;
    }

    resolveType(type) {
        switch (type) {
            case 'success':
            case 'warn':
            case 'error':
                return type;
            case 'info':
            default:
                return 'info';
        }
    }

    resolveText(type) {
        switch (type) {
            case 'error':
                return 'Error';
            case 'success':
                return 'Success';
            case 'info':
                return 'Info';
            case 'warn':
                return 'Warning';
        }

        return 'Info';
    }

    resolveBootstrapType(type) {
        switch (type) {
            case 'warn':
                return 'warning';
            case 'error':
                return 'danger';
            case 'success':
                return 'success';
            case 'info':
            default:
                return 'info';
        }
    }

    resolveIcon(type) {
        var iconClass = `n8go-alert__icon text-${this.resolveBootstrapType(type)}`;

        switch (type) {
            case 'warn':
                return `<svg xmlns="http://www.w3.org/2000/svg" class="${iconClass} icon icon-tabler icon-tabler-alert-triangle" width="20" height="20" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" fill="none"
              stroke-linecap="round" stroke-linejoin="round">
              <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
              <path d="M12 9v2m0 4v.01"></path>
              <path d="M5 19h14a2 2 0 0 0 1.84 -2.75l-7.1 -12.25a2 2 0 0 0 -3.5 0l-7.1 12.25a2 2 0 0 0 1.75 2.75">
              </path>
            </svg>`
                break;
            case 'info':
                return `<svg xmlns="http://www.w3.org/2000/svg" class="${iconClass} icon icon-tabler icon-tabler-info-circle"
                width="20" height="20" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none"
                stroke-linecap="round" stroke-linejoin="round">
                <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
                <circle cx="12" cy="12" r="9"></circle>
                <line x1="12" y1="8" x2="12.01" y2="8"></line>
                <polyline points="11 12 12 12 12 16 13 16"></polyline>
              </svg>`
                break;
            case 'success':
                return `<svg xmlns="http://www.w3.org/2000/svg" class="${iconClass} icon icon-tabler icon-tabler-circle-check" width="20" height="20" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" fill="none"
                stroke-linecap="round" stroke-linejoin="round">
                <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
                 <circle cx="12" cy="12" r="9"></circle>
              <path d="M9 12l2 2l4 -4"></path>
             </svg>`
                break;
            case 'error':
                return `<svg xmlns="http://www.w3.org/2000/svg" class="${iconClass} icon icon-tabler icon-tabler-alert-octagon" width="20" height="20" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" fill="none"
              stroke-linecap="round" stroke-linejoin="round">
              <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
              <path
                d="M8.7 3h6.6c.3 0 .5 .1 .7 .3l4.7 4.7c.2 .2 .3 .4 .3 .7v6.6c0 .3 -.1 .5 -.3 .7l-4.7 4.7c-.2 .2 -.4 .3 -.7 .3h-6.6c-.3 0 -.5 -.1 -.7 -.3l-4.7 -4.7c-.2 -.2 -.3 -.4 -.3 -.7v-6.6c0 -.3 .1 -.5 .3 -.7l4.7 -4.7c.2 -.2 .4 -.3 .7 -.3z">
              </path>
              <line x1="12" y1="8" x2="12" y2="12"></line>
              <line x1="12" y1="16" x2="12.01" y2="16"></line>
            </svg>`
        }

        return '';
    }

    escapeHTML(value) {
        return value
            .replaceAll('&', '&amp;')
            .replaceAll('<', '&lt;')
            .replaceAll('>', '&gt;')
            .replaceAll('"', '&quot;')
            .replaceAll("'", '&#39;');
    }

    connectedCallback() {
        if (!this.rendered) {
            this.render();
            this.rendered = true;
        }
    }

    static get observedAttributes() {
        return ['type', 'message'];
    }

    attributeChangedCallback(name, oldValue, newValue) {
        this.render();
    }

}

if (!customElements.get('n8go-alert')) {
    customElements.define('n8go-alert', Alert);
}
