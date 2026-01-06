// Plausible Analytics helper functions for custom event tracking
const Analytics = {
    // Track search events
    trackSearch: function(query, resultsCount) {
        if (typeof plausible === 'function' && query) {
            plausible('Search', {
                props: {
                    query: query.substring(0, 50),
                    results_count: String(resultsCount),
                    has_results: resultsCount > 0 ? 'true' : 'false'
                }
            });
        }
    },

    // Track filter usage
    trackFilter: function(collection, filterType, filterValue) {
        if (typeof plausible === 'function') {
            plausible('Filter Applied', {
                props: {
                    collection: collection,
                    filter_type: filterType,
                    filter_value: filterValue
                }
            });
        }
    },

    // Track content view (item detail page)
    trackContentView: function(collection, itemSlug, itemTitle) {
        if (typeof plausible === 'function') {
            plausible('Content View', {
                props: {
                    collection: collection,
                    item: itemSlug,
                    title: (itemTitle || '').substring(0, 100)
                }
            });
        }
    },

    // Track collection browse
    trackCollectionBrowse: function(collection, page) {
        if (typeof plausible === 'function') {
            plausible('Collection Browse', {
                props: {
                    collection: collection,
                    page: String(page)
                }
            });
        }
    },

    // Track copy markdown action
    trackCopyMarkdown: function(collection, itemSlug) {
        if (typeof plausible === 'function') {
            plausible('Copy Markdown', {
                props: {
                    collection: collection,
                    item: itemSlug
                }
            });
        }
    }
};

window.Analytics = Analytics;
