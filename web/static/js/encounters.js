// Encounters Calculator JavaScript
// Depends on main.js for shared UI initialization.

function initEncountersPage() {
    initEncounterForm();
    initSteppers();
    initMonsterCart();
    initPickerReset();
    initPickerURLSync();
    syncSourceShort();
    initAutoCalculate();
    initCopyLink();
    initResultShell();
}

function initResultShell() {
    var media = window.matchMedia('(min-width: 900px)');
    document.querySelectorAll('.encounter-result-shell').forEach(function(shell) {
        if (media.matches) shell.open = true;
    });
    if (window.__encounterResultShellBound) return;
    window.__encounterResultShellBound = true;
    media.addEventListener('change', function(evt) {
        document.querySelectorAll('.encounter-result-shell').forEach(function(shell) {
            shell.open = evt.matches;
        });
    });
}

window.initEncountersPage = initEncountersPage;
document.addEventListener('DOMContentLoaded', initEncountersPage);
document.addEventListener('htmx:afterSwap', initEncountersPage);
document.addEventListener('htmx:responseError', function() {
    showToast('Impossibile aggiornare il combattimento. Controlla i dati e riprova.', 'error');
});
document.addEventListener('htmx:sendError', function() {
    showToast('Connessione non disponibile. Riprova tra poco.', 'error');
});
document.addEventListener('htmx:configRequest', function(evt) {
    var form = evt.detail.elt && evt.detail.elt.closest ? evt.detail.elt.closest('.monster-picker-controls') : null;
    if (!form) return;
    var minInput = form.querySelector('[name="min_cr"]');
    var maxInput = form.querySelector('[name="max_cr"]');
    if (!minInput || !maxInput || minInput.value === '' || maxInput.value === '') return;
    if (parseFloat(minInput.value) > parseFloat(maxInput.value)) {
        evt.preventDefault();
        showToast('Il GS minimo non può superare il GS massimo.', 'error');
    }
});

// Toast notification
function showToast(message, type) {
    type = type || 'info';
    var toast = document.createElement('div');
    toast.className = 'flash-message flash-' + type;
    toast.setAttribute('role', type === 'error' ? 'alert' : 'status');
    toast.setAttribute('aria-live', type === 'error' ? 'assertive' : 'polite');
    toast.style.cssText = 'position:fixed;top:20px;right:20px;z-index:1000;min-width:200px;';
    toast.textContent = message;
    document.body.appendChild(toast);
    setTimeout(function() {
        if (toast.parentElement) toast.remove();
    }, 3000);
}

// Generic stepper component
function initSteppers() {
    document.querySelectorAll('.stepper').forEach(function(stepper) {
        if (stepper.dataset.encounterStepperBound === '1') return;
        var input = stepper.querySelector('.stepper-input');
        var decBtn = stepper.querySelector('.stepper-decrement');
        var incBtn = stepper.querySelector('.stepper-increment');
        if (!input || !decBtn || !incBtn) return;
        stepper.dataset.encounterStepperBound = '1';

        var min = parseInt(stepper.dataset.min || input.min || '0', 10);
        var max = parseInt(stepper.dataset.max || input.max || '999', 10);

        function update(delta) {
            var val = parseInt(input.value, 10) || min;
            val = Math.max(min, Math.min(max, val + delta));
            input.value = val;
            decBtn.disabled = val <= min;
            incBtn.disabled = val >= max;
            input.dispatchEvent(new Event('change', { bubbles: true }));
        }

        decBtn.addEventListener('click', function() { update(-1); });
        incBtn.addEventListener('click', function() { update(1); });

        // Validate on direct input
        input.addEventListener('change', function() {
            var val = parseInt(input.value, 10);
            if (isNaN(val)) val = min;
            val = Math.max(min, Math.min(max, val));
            input.value = val;
            decBtn.disabled = val <= min;
            incBtn.disabled = val >= max;
        });

        // Initialize button states
        var initVal = parseInt(input.value, 10) || min;
        decBtn.disabled = initVal <= min;
        incBtn.disabled = initVal >= max;
    });
}

// Debounced auto-calculate: fires custom 'calculate' event that HTMX listens to.
// Each schedule also pushes the share-URL so the address bar tracks party /
// ruleset / difficulty changes without waiting for the calculate response.
function initAutoCalculate() {
    var form = document.getElementById('encounter-form');
    if (!form || typeof htmx === 'undefined') return;
    if (form.dataset.encounterAutoCalculateBound === '1') return;
    form.dataset.encounterAutoCalculateBound = '1';

    var timer = null;
    function scheduleCalculate() {
        clearTimeout(timer);
        syncShareURL();
        timer = setTimeout(function() {
            htmx.trigger(form, 'calculate');
        }, 200);
    }

    form.addEventListener('change', scheduleCalculate);
    form.addEventListener('input', function(evt) {
        if (evt.target.matches('.stepper-input, input[type="number"]')) {
            scheduleCalculate();
        }
    });

}

// Sync the hidden `source_short` input from the checked ruleset radio's
// data-source-short attribute, and propagate changes. Also drives the picker
// source.
function syncSourceShort() {
    var hidden = document.getElementById('source-short');
    var radios = document.querySelectorAll('input[name="ruleset"]');
    if (!hidden || !radios.length) return;

    function update() {
        var checked = document.querySelector('input[name="ruleset"]:checked');
        if (!checked) return;
        var prev = hidden.value;
        var next = checked.dataset.sourceShort || '';
        if (prev && prev !== next && cartQuantities.size > 0) {
            if (!window.confirm('Cambiare edizione rimuoverà tutti i mostri dal combattimento. Continuare?')) {
                radios.forEach(function(radio) {
                    radio.checked = radio.dataset.sourceShort === prev;
                });
                return;
            }
            clearCart();
        }
        hidden.value = next;
        syncDifficultyOptions(checked.value);
        if (prev && prev !== next) refreshMonsterPicker(next);
    }

    if (hidden.dataset.encounterSourceShortBound !== '1') {
        hidden.dataset.encounterSourceShortBound = '1';
        radios.forEach(function(r) { r.addEventListener('change', update); });
    }
    update();
}

function syncDifficultyOptions(ruleset) {
    var select = document.getElementById('encounter-difficulty');
    if (!select || select.dataset.ruleset === ruleset) return;
    var options = ruleset === '2014'
        ? [['Facile', 'Facile'], ['Media', 'Media'], ['Difficile', 'Difficile'], ['Letale', 'Letale']]
        : [['Low', 'Bassa'], ['Moderate', 'Moderata'], ['High', 'Alta']];
    select.innerHTML = '';
    options.forEach(function(option) {
        var element = document.createElement('option');
        element.value = option[0];
        element.textContent = option[1];
        select.appendChild(element);
    });
    select.value = ruleset === '2014' ? 'Media' : 'Moderate';
    select.dataset.ruleset = ruleset;
}

function refreshMonsterPicker(source) {
    var pickerForm = document.querySelector('.monster-picker-controls');
    if (!pickerForm || typeof htmx === 'undefined') return;
    var sourceInput = pickerForm.querySelector('input[name="source"]');
    if (sourceInput) sourceInput.value = source;
    var params = new URLSearchParams(new FormData(pickerForm));
    htmx.ajax('GET', '/combattimenti/monsters?' + params.toString(), {
        target: '#monster-picker-slot',
        swap: 'innerHTML'
    });
}

// Monster cart management: tracks per-(id@source) quantities in a Map and
// re-renders compact hidden <input name="monsters[]"> entries. The server
// accepts "{id}@{source}:qty", which keeps large encounters cheap to update.
var cartQuantities = new Map(); // key "id@source" → integer quantity

function cartKey(id, source) { return id + '@' + source; }

function initMonsterCart() {
    var form = document.getElementById('encounter-form');
    var cartInputs = document.getElementById('cart-inputs');
    if (!form || !cartInputs) return;

    // Hydrate quantities from any pre-existing hidden inputs (e.g. server-side
    // initial render with a non-empty cart).
    rebuildCartFromInputs();

    if (window.__encounterCartClicksBound) return;
    window.__encounterCartClicksBound = true;

    // Delegate clicks: add from picker, step or remove from cart chips.
    document.addEventListener('click', function(evt) {
        var addBtn = evt.target.closest('.monster-picker-row-add');
        if (addBtn) {
            evt.preventDefault();
            var id = addBtn.dataset.monsterId;
            var source = addBtn.dataset.monsterSource;
            if (!id || !source) return;
            adjustCartEntry(id, source, 1);
            var quantity = cartQuantities.get(cartKey(id, source)) || 1;
            addBtn.textContent = 'Aggiunto · ' + quantity;
            addBtn.classList.add('is-added');
            return;
        }

        var stepBtn = evt.target.closest('.result-cart-chip-step');
        if (stepBtn) {
            evt.preventDefault();
            var sid = stepBtn.dataset.cartId;
            var ssrc = stepBtn.dataset.cartSource;
            var delta = parseInt(stepBtn.dataset.cartDelta, 10);
            if (!sid || !ssrc || isNaN(delta)) return;
            adjustCartEntry(sid, ssrc, delta);
            return;
        }

        var removeBtn = evt.target.closest('.result-cart-chip-remove');
        if (removeBtn) {
            evt.preventDefault();
            var rid = removeBtn.dataset.cartId;
            var rsrc = removeBtn.dataset.cartSource;
            if (!rid || !rsrc) return;
            removeCartEntry(rid, rsrc);
        }
    });
}

// adjustCartEntry increments or decrements a chip's quantity. Reaching zero
// removes the chip entirely; clamping at zero keeps the model coherent if a
// stale −1 click arrives.
function adjustCartEntry(id, source, delta) {
    var key = cartKey(id, source);
    var current = cartQuantities.get(key) || 0;
    var next = current + delta;
    if (next <= 0) {
        cartQuantities.delete(key);
    } else {
        cartQuantities.set(key, next);
    }
    renderCartInputs();
    syncShareURL();
    triggerCalculate();
}

function removeCartEntry(id, source) {
    cartQuantities.delete(cartKey(id, source));
    renderCartInputs();
    syncShareURL();
    triggerCalculate();
}

function renderCartInputs() {
    var cartInputs = document.getElementById('cart-inputs');
    if (!cartInputs) return;
    cartInputs.innerHTML = '';
    cartQuantities.forEach(function(qty, key) {
        var input = document.createElement('input');
        input.type = 'hidden';
        input.name = 'monsters[]';
        input.value = qty > 1 ? key + ':' + qty : key;
        cartInputs.appendChild(input);
    });
}

// rebuildCartFromInputs hydrates the in-memory quantity map from whatever
// hidden inputs the server happened to render. Keeps SSR fallback consistent.
function rebuildCartFromInputs() {
    var cartInputs = document.getElementById('cart-inputs');
    cartQuantities = new Map();
    if (!cartInputs) return;
    var inputs = cartInputs.querySelectorAll('input[name="monsters[]"]');
    inputs.forEach(function(input) {
        var raw = input.value;
        var separator = raw.indexOf(':');
        var key = separator === -1 ? raw : raw.slice(0, separator);
        var qty = separator === -1 ? 1 : parseInt(raw.slice(separator + 1), 10);
        if (!key || key.indexOf('@') === -1) return;
        if (isNaN(qty) || qty < 1) return;
        cartQuantities.set(key, (cartQuantities.get(key) || 0) + qty);
    });
}

function clearCart() {
    cartQuantities = new Map();
    var cartInputs = document.getElementById('cart-inputs');
    if (cartInputs) cartInputs.innerHTML = '';
    syncShareURL();
}

function triggerCalculate() {
    var form = document.getElementById('encounter-form');
    if (!form || typeof htmx === 'undefined') return;
    htmx.trigger(form, 'calculate');
}

// --- URL state sync ---------------------------------------------------------
//
// Mirrors the server-side url_state.go encoder so the browser URL stays in
// sync with the form between calculate POSTs (the POST itself sets HX-Push-Url
// which is authoritative — this is a best-effort optimistic update for cart
// add/remove clicks that fire while the next debounced calculate is pending).

function buildShareURL() {
    var form = document.getElementById('encounter-form');
    if (!form) return window.location.pathname;

    var params = new URLSearchParams();

    // ruleset (omit "2024" — the default)
    var ruleset = form.querySelector('input[name="ruleset"]:checked');
    if (ruleset && ruleset.value && ruleset.value !== '2024') {
        params.set('ruleset', ruleset.value);
    }

    var difficulty = form.querySelector('[name="difficulty"]');
    if (difficulty) {
        var defaultDifficulty = ruleset && ruleset.value === '2014' ? 'Media' : 'Moderate';
        if (difficulty.value && difficulty.value !== defaultDifficulty) {
            params.set('diff', difficulty.value);
        }
    }

    // party — comma-separated levels derived from the active panel
    var partyMode = form.querySelector('input[name="party_mode"]:checked');
    var levels = [];
    if (partyMode && partyMode.value === 'different') {
        form.querySelectorAll('#party-different-panel input[name="character_levels"]').forEach(function(input) {
            var n = parseInt(input.value, 10);
            if (!isNaN(n) && n >= 1 && n <= 20) levels.push(n);
        });
    } else {
        var levelInput = form.querySelector('#party-level');
        var countInput = form.querySelector('#party-count');
        var level = levelInput ? parseInt(levelInput.value, 10) : NaN;
        var count = countInput ? parseInt(countInput.value, 10) : NaN;
        if (!isNaN(level) && !isNaN(count)) {
            for (var i = 0; i < count; i++) levels.push(level);
        }
    }
    var partyStr = levels.join(',');
    if (partyStr && partyStr !== '3,3,3,3') {
        params.set('party', partyStr);
    }

    // cart — id@source[:qty], coalescing duplicates
    var cartInputs = form.querySelectorAll('input[name="monsters[]"]');
    if (cartInputs.length) {
        var order = [];
        var counts = {};
        cartInputs.forEach(function(input) {
            var v = (input.value || '').trim();
            if (!v.includes('@')) return;
            var separator = v.indexOf(':');
            var key = separator === -1 ? v : v.slice(0, separator);
            var qty = separator === -1 ? 1 : parseInt(v.slice(separator + 1), 10);
            if (isNaN(qty) || qty < 1) return;
            if (counts[key] === undefined) {
                counts[key] = 0;
                order.push(key);
            }
            counts[key] += qty;
        });
        if (order.length) {
            var parts = order.map(function(key) {
                return counts[key] > 1 ? key + ':' + counts[key] : key;
            });
            params.set('cart', parts.join(','));
        }
    }

    var pickerForm = document.querySelector('.monster-picker-controls');
    if (pickerForm) {
        [['q', 'q'], ['min_cr', 'min_cr'], ['max_cr', 'max_cr'], ['type', 'type'], ['limit', 'limit']].forEach(function(names) {
            var input = pickerForm.querySelector('[name="' + names[0] + '"]');
            var value = input ? String(input.value || '').trim() : '';
            if (value && !(names[1] === 'limit' && value === '20')) params.set(names[1], value);
        });
    }

    var qs = params.toString();
    return qs ? window.location.pathname + '?' + qs : window.location.pathname;
}

// syncShareURL pushes the current form state into the address bar without a
// page reload. Called after every cart mutation; the next HX-Push-Url from
// the server confirms or corrects it.
function syncShareURL() {
    if (typeof window === 'undefined' || !window.history || !window.history.replaceState) return;
    var next = buildShareURL();
    try {
        window.history.replaceState(window.history.state, '', next);
    } catch (e) {
        // replaceState can throw on file:// or sandboxed contexts; ignore.
    }
}

// --- Picker reset -----------------------------------------------------------

// Keeps reset visibility in sync with the stable picker controls, and clears
// those controls before refetching when reset is activated.
function initPickerReset() {
    document.querySelectorAll('form.monster-picker-controls').forEach(function(form) {
        if (form.dataset.encounterPickerResetBound === '1') return;
        form.dataset.encounterPickerResetBound = '1';
        var reset = form.querySelector('.monster-picker-reset');
        if (!reset) return;

        function updateVisibility() {
            var search = form.querySelector('input[type="search"]');
            var hasSearch = search && search.value.trim() !== '';
            var hasNumber = Array.from(form.querySelectorAll('input[type="number"]')).some(function(input) {
                return input.value !== '';
            });
            var type = form.querySelector('select[name="type"]');
            reset.hidden = !(hasSearch || hasNumber || (type && type.value !== ''));
        }

        form.addEventListener('input', updateVisibility);
        form.addEventListener('change', updateVisibility);
        updateVisibility();
    });

    if (window.__encounterPickerResetBound) return;
    window.__encounterPickerResetBound = true;

    document.addEventListener('click', function(evt) {
        var btn = evt.target.closest('.monster-picker-reset');
        if (!btn) return;
        evt.preventDefault();
        var form = btn.closest('form.monster-picker-controls');
        if (!form) return;
        var search = form.querySelector('input[type="search"]');
        if (search) {
            search.value = '';
        }
        form.querySelectorAll('input[type="number"]').forEach(function(input) {
            input.value = '';
        });
        form.querySelectorAll('select').forEach(function(sel) {
            sel.value = '';
        });
        var limit = form.querySelector('input[name="limit"]');
        if (limit) limit.value = '20';
        var source = form.querySelector('input[name="source"]');
        btn.hidden = true;
        refreshMonsterPicker(source ? source.value : '');
    });
}

function initPickerURLSync() {
    if (window.__encounterPickerURLBound) return;
    window.__encounterPickerURLBound = true;
    document.addEventListener('htmx:afterRequest', function(evt) {
        if (evt.detail.elt && evt.detail.elt.closest && evt.detail.elt.closest('.monster-picker')) {
            setTimeout(syncShareURL, 0);
        }
    });
}

// --- Copy link --------------------------------------------------------------

// Delegated handler because the result panel is swapped in by HTMX and the
// button lives inside it. Falls back to a no-op when the clipboard API is
// unavailable (insecure context, ancient browser).
function initCopyLink() {
    if (window.__encounterCopyLinkBound) return;
    window.__encounterCopyLinkBound = true;

    document.addEventListener('click', function(evt) {
        var btn = evt.target.closest('#copy-link-btn');
        if (!btn) return;
        evt.preventDefault();
        copyCurrentURL(btn);
    });
}

function copyCurrentURL(btn) {
    syncShareURL();
    var url = window.location.href;
    var done = function() {
        var success = btn.dataset.copySuccess || 'Copiato!';
        var original = btn.dataset.copyDefault || btn.textContent;
        btn.textContent = success;
        btn.disabled = true;
        setTimeout(function() {
            btn.textContent = original;
            btn.disabled = false;
        }, 1500);
    };
    var fail = function() {
        showToast('Impossibile copiare il link.', 'error');
    };

    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(url).then(done, function() {
            // Fall back to the legacy selection path on permission denial.
            if (!copyWithSelection(url)) fail(); else done();
        });
        return;
    }
    if (copyWithSelection(url)) done(); else fail();
}

// copyWithSelection is the document.execCommand('copy') fallback used when
// the async Clipboard API isn't available (e.g. older browsers, http origins).
function copyWithSelection(text) {
    try {
        var ta = document.createElement('textarea');
        ta.value = text;
        ta.setAttribute('readonly', '');
        ta.style.position = 'fixed';
        ta.style.opacity = '0';
        document.body.appendChild(ta);
        ta.select();
        var ok = document.execCommand('copy');
        document.body.removeChild(ta);
        return ok;
    } catch (e) {
        return false;
    }
}

// Encounter form logic
function initEncounterForm() {
    var form = document.getElementById('encounter-form');
    if (!form) return;
    if (form.dataset.encounterFormBound === '1') return;
    form.dataset.encounterFormBound = '1';

    // Party mode toggle
    var partyModeRadios = document.querySelectorAll('input[name="party_mode"]');
    var partySamePanel = document.getElementById('party-same-panel');
    var partyDifferentPanel = document.getElementById('party-different-panel');

    function updatePartyModeUI() {
        var checked = document.querySelector('input[name="party_mode"]:checked');
        if (!checked) return;
        var isSame = checked.value === 'same';
        if (partySamePanel) partySamePanel.classList.toggle('hidden', !isSame);
        if (partyDifferentPanel) partyDifferentPanel.classList.toggle('hidden', isSame);
    }

    partyModeRadios.forEach(function(radio) {
        radio.addEventListener('change', updatePartyModeUI);
    });

    // Character management for different levels
    var addCharBtn = document.getElementById('add-character');
    var removeCharBtn = document.getElementById('remove-character');
    var charContainer = document.getElementById('character-levels-container');

    if (addCharBtn && charContainer) {
        addCharBtn.addEventListener('click', function() {
            var count = charContainer.children.length;
            if (count < 8) {
                var div = document.createElement('div');
                div.className = 'character-input-group';
                var inputID = 'character-level-' + (count + 1);
                div.innerHTML = '<label for="' + inputID + '">PG ' + (count + 1) + '</label>' +
                    '<input id="' + inputID + '" type="number" name="character_levels" value="3" min="1" max="20" required/>';
                charContainer.appendChild(div);
                div.querySelector('input').dispatchEvent(new Event('change', { bubbles: true }));
                updateCharacterControls();
            }
        });

        if (removeCharBtn) {
            removeCharBtn.addEventListener('click', function() {
                if (charContainer.children.length > 1) {
                    charContainer.removeChild(charContainer.lastElementChild);
                    form.dispatchEvent(new Event('change', { bubbles: true }));
                    updateCharacterControls();
                }
            });
        }

        function updateCharacterControls() {
            var count = charContainer.children.length;
            if (removeCharBtn) removeCharBtn.disabled = count <= 1;
            if (addCharBtn) addCharBtn.disabled = count >= 8;
        }

        updateCharacterControls();
    }

}
