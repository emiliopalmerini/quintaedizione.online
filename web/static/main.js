// Copy markdown functionality with fallback support
function copyMarkdown(btn){
	const section = btn.closest('[data-md-section]') || btn.closest('article') || document;
	const contentEl = section.querySelector('[data-raw]');
	const raw = (contentEl && contentEl.getAttribute('data-raw')) || '';

	if(!raw) {
		showCopyMessage(btn, 'Nessun contenuto da copiare', 'error');
		return;
	}

	// Try modern clipboard API first
	if (navigator.clipboard && navigator.clipboard.writeText) {
		navigator.clipboard.writeText(raw).then(() => {
			showCopyMessage(btn, 'Copiato', 'success');
		}).catch(() => {
			// Fallback to legacy method if modern API fails
			copyToClipboardFallback(raw, btn);
		});
	} else {
		// Use fallback method for browsers without clipboard API
		copyToClipboardFallback(raw, btn);
	}
}

// Fallback copy method using document.execCommand
function copyToClipboardFallback(text, btn) {
	try {
		// Create temporary textarea element
		const textarea = document.createElement('textarea');
		textarea.value = text;
		textarea.style.position = 'fixed';
		textarea.style.left = '-9999px';
		textarea.style.top = '-9999px';
		textarea.style.opacity = '0';
		textarea.setAttribute('readonly', '');
		
		document.body.appendChild(textarea);
		
		// Select and copy text
		textarea.select();
		textarea.setSelectionRange(0, textarea.value.length);
		
		const successful = document.execCommand('copy');
		document.body.removeChild(textarea);
		
		if (successful) {
			showCopyMessage(btn, 'Copiato', 'success');
		} else {
			// Final fallback: select text for manual copy
			showSelectableText(text, btn);
		}
	} catch (err) {
		// Final fallback: show selectable text
		showSelectableText(text, btn);
	}
}

// Show copy message with different types
function showCopyMessage(btn, message, type = 'success') {
	// Remove any existing messages
	const existingMessage = btn.nextElementSibling;
	if (existingMessage && existingMessage.className.includes('flash-message')) {
		existingMessage.remove();
	}
	
	const flash = document.createElement('span');
	flash.className = `flash-message flash-${type}`;
	flash.textContent = message;
	btn.insertAdjacentElement('afterend', flash);
	setTimeout(() => flash.remove(), type === 'error' ? 3000 : 1200);
}

// Final fallback: show text in a modal for manual selection
function showSelectableText(text, btn) {
	const modal = document.createElement('div');
	modal.className = 'copy-text-modal';
	modal.innerHTML = `
		<div class="copy-text-modal-content">
			<div class="copy-text-header">
				<h3>Seleziona e copia il testo</h3>
				<button class="copy-text-close">×</button>
			</div>
			<textarea readonly class="copy-text-area">${text}</textarea>
			<div class="copy-text-footer">
				<small>Seleziona tutto il testo (Ctrl+A) e copialo (Ctrl+C)</small>
			</div>
		</div>
	`;
	
	document.body.appendChild(modal);
	
	// Close button handler
	const closeBtn = modal.querySelector('.copy-text-close');
	closeBtn.addEventListener('click', () => modal.remove());
	
	// Auto-select text and focus textarea
	const textarea = modal.querySelector('.copy-text-area');
	setTimeout(() => {
		textarea.focus();
		textarea.select();
	}, 100);
	
	// Close modal on background click
	modal.addEventListener('click', function(e) {
		if (e.target === modal) {
			modal.remove();
		}
	});
	
	// Close modal on Escape key
	document.addEventListener('keydown', function escapeHandler(e) {
		if (e.key === 'Escape') {
			modal.remove();
			document.removeEventListener('keydown', escapeHandler);
		}
	});
}


// Breadcrumb search functionality
function initBreadcrumbSearch() {
	const searchInput = document.getElementById('bc-search');
	const searchResults = document.getElementById('bc-results');
	
	if (!searchInput || !searchResults) return;
	
	// Show results when typing
	searchInput.addEventListener('input', function() {
		if (this.value.trim().length > 0) {
			searchResults.classList.remove('hidden');
			searchResults.setAttribute('aria-expanded', 'true');
		} else {
			searchResults.classList.add('hidden');
			searchResults.setAttribute('aria-expanded', 'false');
		}
	});
	
	// Hide results when clicking outside
	document.addEventListener('click', function(e) {
		const searchContainer = searchInput.closest('.search-container');
		if (searchContainer && !searchContainer.contains(e.target)) {
			searchResults.classList.add('hidden');
			searchResults.setAttribute('aria-expanded', 'false');
		}
	});
	
	// Handle keyboard navigation in search input
	searchInput.addEventListener('keydown', function(e) {
		const results = searchResults.querySelectorAll('.search-result');

		if (e.key === 'Escape') {
			searchResults.classList.add('hidden');
			searchResults.setAttribute('aria-expanded', 'false');
			this.blur();
		} else if (e.key === 'ArrowDown' && results.length > 0) {
			e.preventDefault();
			results[0].focus();
		} else if (e.key === 'Enter') {
			e.preventDefault();
			// If there are search results, click the first one
			if (results.length > 0) {
				const firstResult = results[0];
				if (firstResult.href) {
					window.location.href = firstResult.href;
				} else {
					firstResult.click();
				}
			}
			// Hide search results after selection
			searchResults.classList.add('hidden');
			searchResults.setAttribute('aria-expanded', 'false');
		}
	});

	// Handle navigation within search results using event delegation
	document.addEventListener('keydown', function(e) {
		// Only handle if we're focused on a search result
		if (!document.activeElement.classList.contains('search-result')) return;

		const results = Array.from(searchResults.querySelectorAll('.search-result[href]'));
		let currentIndex = -1;

		// Find current focused element index
		for (let i = 0; i < results.length; i++) {
			if (results[i] === document.activeElement) {
				currentIndex = i;
				break;
			}
		}

		if (e.key === 'ArrowDown') {
			e.preventDefault();
			const nextIndex = currentIndex < results.length - 1 ? currentIndex + 1 : 0;
			results[nextIndex].focus();
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			if (currentIndex > 0) {
				results[currentIndex - 1].focus();
			} else {
				searchInput.focus();
			}
		} else if (e.key === 'Enter') {
			e.preventDefault();
			const focusedResult = document.activeElement;
			if (focusedResult.href) {
				window.location.href = focusedResult.href;
			} else {
				focusedResult.click();
			}
			searchResults.classList.add('hidden');
			searchResults.setAttribute('aria-expanded', 'false');
		} else if (e.key === 'Escape') {
			searchResults.classList.add('hidden');
			searchResults.setAttribute('aria-expanded', 'false');
			searchInput.focus();
		}
	});
}

// Initialize breadcrumb search on page load and after HTMX swaps
document.addEventListener('DOMContentLoaded', initBreadcrumbSearch);
document.body.addEventListener('htmx:afterSwap', initBreadcrumbSearch);

// Sticky search bar scroll shadow effect
function initStickySearchShadow() {
	const searchBar = document.querySelector('.search-sticky');
	if (!searchBar) return;

	let ticking = false;
	const handleScroll = () => {
		if (!ticking) {
			window.requestAnimationFrame(() => {
				const scrolled = window.scrollY > 10;
				searchBar.classList.toggle('scrolled', scrolled);
				ticking = false;
			});
			ticking = true;
		}
	};

	window.addEventListener('scroll', handleScroll, { passive: true });
}

document.addEventListener('DOMContentLoaded', initStickySearchShadow);
document.body.addEventListener('htmx:afterSwap', initStickySearchShadow);

// Global search ESC handling
function initGlobalSearch() {
	const searchInput = document.getElementById('global-search');
	const searchResults = document.getElementById('search-results');
	const searchForm = document.getElementById('search-form');
	const searchCloseBtn = document.getElementById('search-close-btn');

	if (!searchInput) return;

	// Prevent HTMX request with less than 2 characters
	searchForm.addEventListener('htmx:beforeRequest', function(evt) {
		if (searchInput.value.trim().length < 2) {
			evt.detail.cancel = true;
			if (searchResults) {
				searchResults.innerHTML = '';
			}
		}
	});

	// ESC key closes dropdown and clears search
	searchInput.addEventListener('keydown', function(e) {
		if (e.key === 'Escape') {
			e.preventDefault();
			if (searchResults) {
				searchResults.innerHTML = '';
			}
			if (searchForm) {
				searchForm.reset();
			}
			this.blur();
		}
	});

	// Click outside closes dropdown
	document.addEventListener('click', function(e) {
		if (searchInput && searchResults && !searchInput.closest('.search-container-wrapper').contains(e.target)) {
			searchResults.innerHTML = '';
		}
	});

	// Close button clears and closes
	if (searchCloseBtn) {
		searchCloseBtn.addEventListener('click', function() {
			if (searchResults) {
				searchResults.innerHTML = '';
			}
			searchInput.focus();
		});
	}
}

document.addEventListener('DOMContentLoaded', initGlobalSearch);
document.body.addEventListener('htmx:afterSwap', initGlobalSearch);

// Back button handler
function initBackButton() {
	const backBtn = document.getElementById('back-btn');
	if (backBtn) {
		backBtn.addEventListener('click', () => history.back());
	}
}

// Copy markdown button handler
function initCopyMarkdownButton() {
	const copyBtn = document.getElementById('copy-markdown-btn');
	if (copyBtn) {
		copyBtn.addEventListener('click', function() {
			copyMarkdown(this);
		});
	}
}

// Prevent form submission for search form
function initSearchFormHandler() {
	const searchForm = document.getElementById('search-form');
	if (searchForm) {
		searchForm.addEventListener('submit', (e) => {
			e.preventDefault();
			return false;
		});
	}
}

// Glossary tooltip functionality
function initGlossaryTooltips() {
	let activeTooltip = null;
	let hideTimeout = null;

	function removeTooltip() {
		if (activeTooltip) {
			activeTooltip.remove();
			activeTooltip = null;
		}
	}

	function showTooltip(term) {
		removeTooltip();
		clearTimeout(hideTimeout);

		const def = term.getAttribute('data-term-def');
		const id = term.getAttribute('data-term-id');
		const cat = term.getAttribute('data-term-cat');

		const tooltip = document.createElement('div');
		tooltip.className = 'glossary-tooltip';

		let html = '';
		if (cat) {
			html += '<div class="glossary-tooltip-cat">' + cat + '</div>';
		}
		html += '<div class="glossary-tooltip-def">' + def + '</div>';
		html += '<a href="/glossario/' + id + '" class="glossary-tooltip-link">Vedi nel glossario →</a>';
		tooltip.innerHTML = html;

		document.body.appendChild(tooltip);
		activeTooltip = tooltip;

		// Position tooltip
		const rect = term.getBoundingClientRect();
		const tooltipRect = tooltip.getBoundingClientRect();

		let top = rect.bottom + window.scrollY + 8;
		let left = rect.left + window.scrollX;

		// Adjust if overflowing right edge
		if (left + tooltipRect.width > window.innerWidth - 16) {
			left = window.innerWidth - tooltipRect.width - 16;
		}
		// Adjust if overflowing left edge
		if (left < 16) {
			left = 16;
		}
		// Show above if overflowing bottom
		if (rect.bottom + tooltipRect.height + 8 > window.innerHeight) {
			top = rect.top + window.scrollY - tooltipRect.height - 8;
		}

		tooltip.style.top = top + 'px';
		tooltip.style.left = left + 'px';

		// Allow moving cursor to tooltip
		tooltip.addEventListener('mouseenter', function() {
			clearTimeout(hideTimeout);
		});
		tooltip.addEventListener('mouseleave', function() {
			hideTimeout = setTimeout(removeTooltip, 150);
		});
	}

	// Event delegation on document body
	document.body.addEventListener('mouseenter', function(e) {
		if (e.target.classList && e.target.classList.contains('glossary-term')) {
			showTooltip(e.target);
		}
	}, true);

	document.body.addEventListener('mouseleave', function(e) {
		if (e.target.classList && e.target.classList.contains('glossary-term')) {
			hideTimeout = setTimeout(removeTooltip, 150);
		}
	}, true);

	// Touch support: tap to toggle
	document.body.addEventListener('click', function(e) {
		if (e.target.classList && e.target.classList.contains('glossary-term')) {
			e.preventDefault();
			if (activeTooltip) {
				removeTooltip();
			} else {
				showTooltip(e.target);
			}
		} else if (activeTooltip && !activeTooltip.contains(e.target)) {
			removeTooltip();
		}
	});

	// Escape key dismisses tooltip
	document.addEventListener('keydown', function(e) {
		if (e.key === 'Escape' && activeTooltip) {
			removeTooltip();
		}
	});
}

document.addEventListener('DOMContentLoaded', () => {
	initBackButton();
	initCopyMarkdownButton();
	initSearchFormHandler();
	initGlossaryTooltips();
});

document.body.addEventListener('htmx:afterSwap', () => {
	initBackButton();
	initCopyMarkdownButton();
	initSearchFormHandler();
});
