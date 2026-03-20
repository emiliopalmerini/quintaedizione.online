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

// Global search (desktop inline dropdown)
function initGlobalSearch() {
	var searchInput = document.getElementById('global-search');
	var searchResults = document.getElementById('search-results');
	var searchForm = document.getElementById('search-form');
	var searchCloseBtn = document.getElementById('search-close-btn');
	var desktopCollection = document.getElementById('desktop-collection');

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

	// Enter key navigates to full search page
	searchInput.addEventListener('keydown', function(e) {
		if (e.key === 'Enter') {
			e.preventDefault();
			var q = searchInput.value.trim();
			if (q.length >= 2) {
				window.location.href = '/srd/search?q=' + encodeURIComponent(q);
			}
		}
		// ESC key closes dropdown and clears search
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
			if (desktopCollection) {
				desktopCollection.value = '';
			}
			// Reset desktop chips
			var chips = document.querySelectorAll('.desktop-chip');
			chips.forEach(function(c) {
				c.classList.toggle('active', c.getAttribute('data-collection') === '');
			});
			searchInput.focus();
		});
	}

	// Desktop filter chip selection
	initDesktopChips(searchInput, desktopCollection);
}

function initDesktopChips(searchInput, desktopCollection) {
	var chips = document.querySelectorAll('.desktop-chip');
	if (!chips.length || !desktopCollection) return;

	chips.forEach(function(chip) {
		chip.addEventListener('click', function() {
			chips.forEach(function(c) { c.classList.remove('active'); });
			chip.classList.add('active');
			desktopCollection.value = chip.getAttribute('data-collection');
			// Re-trigger search with new collection filter
			if (searchInput && searchInput.value.trim().length >= 2) {
				htmx.trigger(searchInput, 'search');
			}
		});
	});
}

// Mobile search overlay (full-screen)
function initSearchOverlay() {
	var overlay = document.getElementById('search-overlay');
	var heroInput = document.getElementById('global-search');
	if (!overlay || !heroInput) return;

	var overlayInput = document.getElementById('overlay-search-input');
	var overlayResults = document.getElementById('overlay-results');
	var overlayCollection = document.getElementById('overlay-collection');
	var overlayForm = document.getElementById('overlay-search-form');
	var closeBtn = overlay.querySelector('.search-overlay-close');
	var chips = overlay.querySelectorAll('.search-overlay-chip');
	var isMobile = window.matchMedia('(max-width: 640px)');

	function openOverlay() {
		if (!isMobile.matches) return;
		overlay.classList.add('open');
		document.body.style.overflow = 'hidden';
		// Sync text from hero input
		if (heroInput.value) {
			overlayInput.value = heroInput.value;
		}
		setTimeout(function() {
			overlayInput.focus();
			// Trigger search if text was synced
			if (overlayInput.value.trim().length >= 2) {
				htmx.trigger(overlayInput, 'keyup');
			}
		}, 50);
	}

	function closeOverlay() {
		overlay.classList.remove('open');
		document.body.style.overflow = '';
		overlayResults.innerHTML = '';
		overlayInput.value = '';
		overlayCollection.value = '';
		// Reset chips
		chips.forEach(function(c) {
			c.classList.toggle('active', c.getAttribute('data-collection') === '');
		});
	}

	// Open overlay when hero input is focused on mobile
	heroInput.addEventListener('focus', function() {
		if (isMobile.matches) {
			heroInput.blur();
			openOverlay();
		}
	});

	// Close button
	if (closeBtn) {
		closeBtn.addEventListener('click', closeOverlay);
	}

	// ESC closes overlay
	document.addEventListener('keydown', function(e) {
		if (e.key === 'Escape' && overlay.classList.contains('open')) {
			closeOverlay();
		}
	});

	// Prevent HTMX request with less than 2 characters
	if (overlayForm) {
		overlayForm.addEventListener('htmx:beforeRequest', function(evt) {
			if (overlayInput.value.trim().length < 2) {
				evt.detail.cancel = true;
				overlayResults.innerHTML = '';
			}
		});
	}

	// Collection chip selection
	chips.forEach(function(chip) {
		chip.addEventListener('click', function() {
			chips.forEach(function(c) { c.classList.remove('active'); });
			chip.classList.add('active');
			overlayCollection.value = chip.getAttribute('data-collection');
			// Re-trigger search with new collection filter via custom event
			if (overlayInput.value.trim().length >= 2) {
				htmx.trigger(overlayInput, 'search');
			}
		});
	});
}

document.addEventListener('DOMContentLoaded', function() {
	initGlobalSearch();
	initSearchOverlay();
});
document.body.addEventListener('htmx:afterSwap', function() {
	initGlobalSearch();
	initSearchOverlay();
});

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

// Mobile filter overlay (full-screen, D&D Beyond style)
function initFilterOverlay() {
	var overlay = document.getElementById('filter-overlay');
	var toggleBtn = document.getElementById('filter-toggle-btn');
	if (!overlay || !toggleBtn) return;

	var filtersChanged = false;

	function openOverlay() {
		filtersChanged = false;
		overlay.classList.add('open');
		overlay.classList.remove('detail');
		overlay.querySelectorAll('.filter-overlay-detail').forEach(function(d) {
			d.classList.remove('active');
		});
		document.body.style.overflow = 'hidden';
	}

	function closeOverlay() {
		overlay.classList.remove('open', 'detail');
		overlay.querySelectorAll('.filter-overlay-detail').forEach(function(d) {
			d.classList.remove('active');
		});
		document.body.style.overflow = '';
		if (filtersChanged) {
			var firstInput = document.querySelector('#search-form .filter-multiselect input[type="hidden"]');
			if (firstInput) {
				firstInput.dispatchEvent(new Event('change', { bubbles: true }));
			}
		}
	}

	function showDetail(filterName) {
		overlay.classList.add('detail');
		overlay.querySelectorAll('.filter-overlay-detail').forEach(function(d) {
			d.classList.toggle('active', d.getAttribute('data-filter-detail') === filterName);
		});
	}

	function showCategories() {
		overlay.classList.remove('detail');
		overlay.querySelectorAll('.filter-overlay-detail').forEach(function(d) {
			d.classList.remove('active');
		});
	}

	function syncOverlayFromForm() {
		document.querySelectorAll('#search-form .filter-multiselect').forEach(function(container) {
			var filterName = container.getAttribute('data-filter-name');
			var hiddenInput = container.querySelector('input[type="hidden"]');
			if (!hiddenInput) return;
			var currentValues = hiddenInput.value ? hiddenInput.value.split(',') : [];
			// Sync overlay checkboxes
			overlay.querySelectorAll('input[type="checkbox"][data-filter-name="' + filterName + '"]').forEach(function(cb) {
				cb.checked = currentValues.indexOf(cb.value) !== -1;
			});
			// Update category count badge
			var categoryEl = overlay.querySelector('.filter-overlay-category[data-filter-name="' + filterName + '"]');
			if (categoryEl) {
				var countEl = categoryEl.querySelector('.filter-overlay-category-count');
				if (currentValues.length > 0 && currentValues[0] !== '') {
					if (!countEl) {
						countEl = document.createElement('span');
						countEl.className = 'filter-overlay-category-count';
						categoryEl.querySelector('.filter-overlay-category-name').insertAdjacentElement('afterend', countEl);
					}
					countEl.textContent = currentValues.length;
				} else if (countEl) {
					countEl.remove();
				}
			}
		});
	}

	function syncFormFromOverlay(filterName) {
		var container = document.querySelector('#search-form .filter-multiselect[data-filter-name="' + filterName + '"]');
		if (!container) return;
		var hiddenInput = container.querySelector('input[type="hidden"]');
		if (!hiddenInput) return;
		var checked = [];
		overlay.querySelectorAll('input[type="checkbox"][data-filter-name="' + filterName + '"]:checked').forEach(function(cb) {
			checked.push(cb.value);
		});
		hiddenInput.value = checked.join(',');
		// Also sync the desktop dropdown checkboxes
		var dropdown = container.querySelector('.filter-multiselect-dropdown');
		if (dropdown) {
			dropdown.querySelectorAll('input[type="checkbox"]').forEach(function(cb) {
				cb.checked = checked.indexOf(cb.value) !== -1;
			});
		}
	}

	// Open overlay on mobile, toggle filter panel on desktop
	toggleBtn.addEventListener('click', function() {
		if (window.innerWidth <= 640) {
			syncOverlayFromForm();
			openOverlay();
		} else {
			var panel = document.getElementById('filter-panel');
			if (panel) {
				panel.classList.toggle('open');
			}
		}
	});

	// Close button
	overlay.querySelectorAll('.filter-overlay-close').forEach(function(btn) {
		btn.addEventListener('click', closeOverlay);
	});

	// Clear all
	overlay.querySelectorAll('.filter-overlay-clear').forEach(function(btn) {
		btn.addEventListener('click', function() {
			overlay.querySelectorAll('input[type="checkbox"]').forEach(function(cb) {
				cb.checked = false;
			});
			// Sync all filters to form
			document.querySelectorAll('#search-form .filter-multiselect').forEach(function(container) {
				var filterName = container.getAttribute('data-filter-name');
				var hiddenInput = container.querySelector('input[type="hidden"]');
				if (hiddenInput) hiddenInput.value = '';
			});
			filtersChanged = true;
			closeOverlay();
		});
	});

	// Category click → show detail
	overlay.querySelectorAll('.filter-overlay-category').forEach(function(cat) {
		cat.addEventListener('click', function() {
			showDetail(this.getAttribute('data-filter-name'));
		});
	});

	// Back button → show categories
	overlay.querySelectorAll('.filter-overlay-back').forEach(function(btn) {
		btn.addEventListener('click', showCategories);
	});

	// Checkbox changes in overlay → sync to form
	overlay.querySelectorAll('input[type="checkbox"]').forEach(function(cb) {
		cb.addEventListener('change', function() {
			var filterName = this.getAttribute('data-filter-name');
			syncFormFromOverlay(filterName);
			filtersChanged = true;
			// Update category count badge
			var categoryEl = overlay.querySelector('.filter-overlay-category[data-filter-name="' + filterName + '"]');
			if (categoryEl) {
				var checked = overlay.querySelectorAll('input[type="checkbox"][data-filter-name="' + filterName + '"]:checked');
				var countEl = categoryEl.querySelector('.filter-overlay-category-count');
				if (checked.length > 0) {
					if (!countEl) {
						countEl = document.createElement('span');
						countEl.className = 'filter-overlay-category-count';
						categoryEl.querySelector('.filter-overlay-category-name').insertAdjacentElement('afterend', countEl);
					}
					countEl.textContent = checked.length;
				} else if (countEl) {
					countEl.remove();
				}
			}
		});
	});

	// Escape key closes overlay
	document.addEventListener('keydown', function(e) {
		if (e.key === 'Escape' && overlay.classList.contains('open')) {
			closeOverlay();
		}
	});
}

// Multi-select filter dropdowns
function initFilterMultiselect() {
	document.querySelectorAll('.filter-multiselect').forEach(function(container) {
		var btn = container.querySelector('.filter-multiselect-btn');
		var dropdown = container.querySelector('.filter-multiselect-dropdown');
		var hiddenInput = container.querySelector('input[type="hidden"]');
		if (!btn || !dropdown || !hiddenInput) return;

		btn.addEventListener('click', function(e) {
			e.preventDefault();
			e.stopPropagation();
			// Close other open dropdowns
			document.querySelectorAll('.filter-multiselect.open').forEach(function(other) {
				if (other !== container) other.classList.remove('open');
			});
			container.classList.toggle('open');
		});

		dropdown.querySelectorAll('input[type="checkbox"]').forEach(function(cb) {
			cb.addEventListener('change', function() {
				var checked = [];
				dropdown.querySelectorAll('input[type="checkbox"]:checked').forEach(function(c) {
					checked.push(c.value);
				});
				hiddenInput.value = checked.join(',');
				hiddenInput.dispatchEvent(new Event('change', { bubbles: true }));
			});
		});
	});

	// Close dropdowns when clicking outside
	document.addEventListener('click', function(e) {
		if (!e.target.closest('.filter-multiselect')) {
			document.querySelectorAll('.filter-multiselect.open').forEach(function(el) {
				el.classList.remove('open');
			});
		}
	});
}

// Quick filter chips: toggle chip → sync with dropdown filter → trigger HTMX
function initQuickFilterChips() {
	document.querySelectorAll('.quick-filter-bar[data-filter-name]').forEach(function(bar) {
		var filterName = bar.getAttribute('data-filter-name');
		bar.querySelectorAll('.quick-filter-chip[data-values]').forEach(function(chip) {
			chip.addEventListener('click', function() {
				var wasActive = chip.classList.contains('active');
				var chipValues = chip.getAttribute('data-values').split(',');

				// Find the corresponding hidden input in the search form
				var container = document.querySelector('.filter-multiselect[data-filter-name="' + filterName + '"]');
				if (!container) return;
				var hiddenInput = container.querySelector('input[type="hidden"]');
				if (!hiddenInput) return;

				var currentValues = hiddenInput.value ? hiddenInput.value.split(',').filter(function(v) { return v !== ''; }) : [];

				if (wasActive) {
					// Deactivate: remove chip values from filter
					currentValues = currentValues.filter(function(v) {
						return chipValues.indexOf(v) === -1;
					});
				} else {
					// Activate: add chip values to filter (avoid duplicates)
					chipValues.forEach(function(v) {
						if (currentValues.indexOf(v) === -1) {
							currentValues.push(v);
						}
					});
				}

				hiddenInput.value = currentValues.join(',');

				// Sync desktop dropdown checkboxes
				var dropdown = container.querySelector('.filter-multiselect-dropdown');
				if (dropdown) {
					dropdown.querySelectorAll('input[type="checkbox"]').forEach(function(cb) {
						cb.checked = currentValues.indexOf(cb.value) !== -1;
					});
				}

				// Sync overlay checkboxes
				var overlay = document.getElementById('filter-overlay');
				if (overlay) {
					overlay.querySelectorAll('input[type="checkbox"][data-filter-name="' + filterName + '"]').forEach(function(cb) {
						cb.checked = currentValues.indexOf(cb.value) !== -1;
					});
				}

				hiddenInput.dispatchEvent(new Event('change', { bubbles: true }));
			});
		});
	});
}

// Mappe tag picker: search + multi-select chips, works after HTMX swaps
function initMappeFilterChips() {
	var picker = document.getElementById('mappe-tag-picker');
	if (!picker) return;

	var hiddenInput = document.getElementById('mappe-tag');
	var searchInput = document.getElementById('mappe-tag-search');
	var optionsContainer = picker.querySelector('.tag-picker-options');
	if (!hiddenInput || !searchInput || !optionsContainer) return;

	function getActiveTags() {
		return hiddenInput.value ? hiddenInput.value.split(',').filter(Boolean) : [];
	}

	function setActiveTags(tags) {
		hiddenInput.value = tags.join(',');
	}

	function rebuildSelected() {
		var existing = picker.querySelector('.tag-picker-selected');
		if (existing) existing.remove();

		var tags = getActiveTags();
		if (tags.length === 0) return;

		var container = document.createElement('div');
		container.className = 'tag-picker-selected';

		tags.forEach(function(tag) {
			var btn = document.createElement('button');
			btn.type = 'button';
			btn.className = 'tag-picker-active';
			btn.setAttribute('data-value', tag);
			btn.innerHTML = tag + ' <span class="tag-picker-remove" aria-hidden="true">&times;</span>';
			btn.addEventListener('click', function() {
				removeTag(tag);
			});
			container.appendChild(btn);
		});

		picker.insertBefore(container, searchInput);
	}

	function addTag(tag) {
		var tags = getActiveTags();
		if (tags.indexOf(tag) === -1) {
			tags.push(tag);
			setActiveTags(tags);
		}
		// Hide the chip from available options
		optionsContainer.querySelectorAll('.quick-filter-chip[data-value]').forEach(function(chip) {
			if (chip.getAttribute('data-value') === tag) {
				chip.style.display = 'none';
			}
		});
		searchInput.value = '';
		filterOptions('');
		rebuildSelected();
		htmx.trigger(document.getElementById('mappe-filter-form'), 'filter-changed');
	}

	function removeTag(tag) {
		var tags = getActiveTags().filter(function(t) { return t !== tag; });
		setActiveTags(tags);
		// Show the chip in available options again
		optionsContainer.querySelectorAll('.quick-filter-chip[data-value]').forEach(function(chip) {
			if (chip.getAttribute('data-value') === tag) {
				chip.style.display = '';
			}
		});
		rebuildSelected();
		htmx.trigger(document.getElementById('mappe-filter-form'), 'filter-changed');
	}

	function filterOptions(query) {
		var q = query.toLowerCase();
		optionsContainer.querySelectorAll('.quick-filter-chip[data-value]').forEach(function(chip) {
			var value = chip.getAttribute('data-value');
			var activeTags = getActiveTags();
			if (activeTags.indexOf(value) !== -1) {
				chip.style.display = 'none';
			} else if (q && value.toLowerCase().indexOf(q) === -1) {
				chip.style.display = 'none';
			} else {
				chip.style.display = '';
			}
		});
	}

	// Bind search input
	searchInput.addEventListener('input', function() {
		filterOptions(this.value);
	});

	// Bind available tag chips
	optionsContainer.querySelectorAll('.quick-filter-chip[data-value]').forEach(function(chip) {
		chip.addEventListener('click', function() {
			addTag(chip.getAttribute('data-value'));
		});
	});

	// Bind existing active tag buttons (server-rendered)
	picker.querySelectorAll('.tag-picker-active[data-value]').forEach(function(btn) {
		btn.addEventListener('click', function() {
			removeTag(btn.getAttribute('data-value'));
		});
	});
}

// Filter chips: remove individual filter value or clear all
function initFilterChips() {
	document.querySelectorAll('.filter-chip').forEach(function(chip) {
		chip.addEventListener('click', function() {
			var filterName = this.getAttribute('data-filter-name');
			var filterValue = this.getAttribute('data-filter-value');
			var container = document.querySelector('.filter-multiselect[data-filter-name="' + filterName + '"]');
			if (!container) return;

			var hiddenInput = container.querySelector('input[type="hidden"]');
			if (!hiddenInput) return;

			// Remove this specific value from the hidden input
			var currentValues = hiddenInput.value.split(',').filter(function(v) {
				return v !== '' && v !== filterValue;
			});
			hiddenInput.value = currentValues.join(',');

			// Uncheck the corresponding checkbox in desktop dropdown
			var cb = container.querySelector('input[type="checkbox"][value="' + filterValue + '"]');
			if (cb) cb.checked = false;

			// Also uncheck in overlay
			var overlayCb = document.querySelector('#filter-overlay input[type="checkbox"][data-filter-name="' + filterName + '"][value="' + filterValue + '"]');
			if (overlayCb) overlayCb.checked = false;

			hiddenInput.dispatchEvent(new Event('change', { bubbles: true }));
		});
	});

	var clearAllBtn = document.getElementById('clear-all-filters');
	if (clearAllBtn) {
		clearAllBtn.addEventListener('click', function() {
			var form = document.getElementById('search-form');
			if (!form) return;

			form.querySelectorAll('.filter-multiselect input[type="hidden"]').forEach(function(input) {
				input.value = '';
			});
			form.querySelectorAll('.filter-multiselect input[type="checkbox"]').forEach(function(cb) {
				cb.checked = false;
			});

			// Also clear overlay checkboxes
			var overlay = document.getElementById('filter-overlay');
			if (overlay) {
				overlay.querySelectorAll('input[type="checkbox"]').forEach(function(cb) {
					cb.checked = false;
				});
			}

			// Trigger a change to fire HTMX
			var firstInput = form.querySelector('.filter-multiselect input[type="hidden"]');
			if (firstInput) {
				firstInput.dispatchEvent(new Event('change', { bubbles: true }));
			}
		});
	}
}

// Theme toggle (both desktop and mobile buttons)
function initThemeToggle() {
	var btns = [document.getElementById('theme-toggle'), document.getElementById('theme-toggle-mobile')];
	btns.forEach(function(btn) {
		if (!btn) return;
		btn.addEventListener('click', function() {
			var d = document.documentElement;
			var isDark = d.classList.toggle('dark');
			localStorage.setItem('theme', isDark ? 'dark' : 'light');
		});
	});
}

// Hamburger menu toggle
function initHamburgerMenu() {
	var btn = document.getElementById('nav-hamburger');
	var menu = document.getElementById('nav-menu');
	if (!btn || !menu) return;

	btn.addEventListener('click', function() {
		var expanded = btn.getAttribute('aria-expanded') === 'true';
		btn.setAttribute('aria-expanded', String(!expanded));
		menu.classList.toggle('open');
	});

	// Close menu when a nav link is clicked
	menu.querySelectorAll('.site-nav-link').forEach(function(link) {
		link.addEventListener('click', function() {
			btn.setAttribute('aria-expanded', 'false');
			menu.classList.remove('open');
		});
	});

	// Close menu when clicking outside
	document.addEventListener('click', function(e) {
		if (!btn.contains(e.target) && !menu.contains(e.target)) {
			btn.setAttribute('aria-expanded', 'false');
			menu.classList.remove('open');
		}
	});
}

// Patreon banner dismiss
function initPatreonBanner() {
	var banner = document.getElementById('patreon-banner');
	if (!banner) return;
	if (localStorage.getItem('patreon-banner-dismissed')) {
		banner.classList.add('patreon-banner--hidden');
		return;
	}
	var dismissBtn = banner.querySelector('.patreon-banner-dismiss');
	if (dismissBtn) {
		dismissBtn.addEventListener('click', function() {
			banner.classList.add('patreon-banner--hidden');
			localStorage.setItem('patreon-banner-dismissed', '1');
		});
	}
}

// Keyboard navigation: arrow keys for prev/next item
function initItemKeyboardNav() {
	var prevBtn = document.querySelector('.item-nav a[aria-label="Precedente"]');
	var nextBtn = document.querySelector('.item-nav a[aria-label="Successivo"]');
	if (!prevBtn && !nextBtn) return;

	document.addEventListener('keydown', function handleItemNav(e) {
		// Skip when user is typing
		var tag = document.activeElement.tagName;
		if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || document.activeElement.isContentEditable) return;

		if (e.altKey || e.ctrlKey || e.metaKey || e.shiftKey) return;

		if (e.key === 'ArrowLeft' && prevBtn) {
			e.preventDefault();
			prevBtn.click();
		} else if (e.key === 'ArrowRight' && nextBtn) {
			e.preventDefault();
			nextBtn.click();
		}
	});
}

// Auto-generate anchor links on headings inside .item-content
function initHeadingAnchors() {
	var container = document.querySelector('.item-content');
	if (!container) return;

	container.querySelectorAll('h2, h3').forEach(function(heading) {
		if (heading.id) return; // already has an anchor

		// Generate id from text content
		var text = heading.textContent.trim();
		var id = text.toLowerCase()
			.replace(/[^\w\s-]/g, '')
			.replace(/\s+/g, '-')
			.replace(/-+/g, '-');
		if (!id) return;

		heading.id = id;
		heading.style.position = 'relative';

		var anchor = document.createElement('a');
		anchor.href = '#' + id;
		anchor.className = 'heading-anchor';
		anchor.setAttribute('aria-label', 'Link a questa sezione');
		anchor.textContent = '#';
		heading.appendChild(anchor);
	});
}

document.addEventListener('DOMContentLoaded', () => {
	initThemeToggle();
	initHamburgerMenu();
	initPatreonBanner();
	initBackButton();
	initCopyMarkdownButton();
	initSearchFormHandler();
	initGlossaryTooltips();
	initFilterOverlay();
	initFilterMultiselect();
	initFilterChips();
	initQuickFilterChips();
	initMappeFilterChips();
	initItemKeyboardNav();
	initHeadingAnchors();
});

document.body.addEventListener('htmx:afterSwap', function(evt) {
	initBackButton();
	initCopyMarkdownButton();
	initSearchFormHandler();
	initFilterOverlay();
	initFilterMultiselect();
	initFilterChips();
	initQuickFilterChips();
	initMappeFilterChips();
	initItemKeyboardNav();
	initHeadingAnchors();

	// Scroll to top of results when rows are updated (filter/pagination change)
	// Skip on history restore (browser back button)
	if (evt.detail && evt.detail.target && evt.detail.target.id === 'rows' && !evt.detail.isHistoryRestoreRequest) {
		var rows = document.getElementById('rows');
		if (rows) {
			rows.scrollIntoView({ behavior: 'smooth', block: 'start' });
		}
	}
});
