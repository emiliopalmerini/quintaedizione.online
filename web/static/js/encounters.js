// Encounters Calculator JavaScript
// Depends on main.js for: initThemeToggle, initPatreonBanner

document.addEventListener('DOMContentLoaded', function() {
    initEncounterForm();
    initSteppers();
    initAutoCalculate();
    initMonsterCart();
    syncSourceShort();
    initCopyLink();

    // Configure HTMX for encounters
    if (typeof htmx !== 'undefined') {
        htmx.config.requestClass = 'loading';
        htmx.config.historyEnabled = true;
    }

    // HTMX loading indicators
    document.addEventListener('htmx:beforeRequest', function(evt) {
        var target = evt.target;
        if (target.classList.contains('btn')) {
            target.style.opacity = '0.7';
            target.style.pointerEvents = 'none';
        }
    });

    document.addEventListener('htmx:afterRequest', function(evt) {
        var target = evt.target;
        if (target.classList.contains('btn')) {
            target.style.opacity = '1';
            target.style.pointerEvents = 'auto';
        }
    });

    document.addEventListener('htmx:responseError', function() {
        showToast('Errore nel caricamento. Riprova.', 'error');
    });
});

// Toast notification
function showToast(message, type) {
    type = type || 'info';
    var toast = document.createElement('div');
    toast.className = 'flash-message flash-' + type;
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
        var input = stepper.querySelector('.stepper-input');
        var decBtn = stepper.querySelector('.stepper-decrement');
        var incBtn = stepper.querySelector('.stepper-increment');
        if (!input || !decBtn || !incBtn) return;

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

    // Calculate on page load with default values
    setTimeout(scheduleCalculate, 100);
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
        hidden.value = next;
        if (prev && prev !== next) {
            clearCart();
        }
    }

    radios.forEach(function(r) { r.addEventListener('change', update); });
    update();
}

// Monster cart management: tracks per-(id@source) quantities in a Map and
// re-renders hidden <input name="monsters[]"> entries (one per unit) so the
// server-side ParseCartRefs (which folds duplicates) sees the right count.
//
// Wire format stays "{id}@{source}" repeated; the server folds duplicates by
// counting occurrences, so this client only needs to emit qty copies per chip.
var cartQuantities = new Map(); // key "id@source" → integer quantity

function cartKey(id, source) { return id + '@' + source; }

function initMonsterCart() {
    var form = document.getElementById('encounter-form');
    var cartInputs = document.getElementById('cart-inputs');
    if (!form || !cartInputs) return;

    // Hydrate quantities from any pre-existing hidden inputs (e.g. server-side
    // initial render with a non-empty cart).
    rebuildCartFromInputs();

    // Delegate clicks: add from picker, step or remove from cart chips.
    document.body.addEventListener('click', function(evt) {
        var addBtn = evt.target.closest('.monster-picker-row-add');
        if (addBtn) {
            evt.preventDefault();
            var id = addBtn.dataset.monsterId;
            var source = addBtn.dataset.monsterSource;
            if (!id || !source) return;
            adjustCartEntry(id, source, 1);
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

// renderCartInputs writes one hidden <input> per unit so the server's
// ParseCartRefs (occurrence-count folding) stays compatible with the legacy
// wire format.
function renderCartInputs() {
    var cartInputs = document.getElementById('cart-inputs');
    if (!cartInputs) return;
    cartInputs.innerHTML = '';
    cartQuantities.forEach(function(qty, key) {
        for (var i = 0; i < qty; i++) {
            var input = document.createElement('input');
            input.type = 'hidden';
            input.name = 'monsters[]';
            input.value = key;
            cartInputs.appendChild(input);
        }
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
        var key = input.value;
        if (!key || key.indexOf('@') === -1) return;
        cartQuantities.set(key, (cartQuantities.get(key) || 0) + 1);
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

    // difficulty — only emit when not the ruleset's default
    var rulesetVal = (ruleset && ruleset.value) || '2024';
    var defaultDiff = rulesetVal === '2014' ? 'Media' : 'Moderate';
    var diff = '';
    if (rulesetVal === '2014') {
        var d2014 = form.querySelector('select[name="difficulty_2014"]');
        if (d2014) diff = d2014.value;
    } else {
        var d2024 = form.querySelector('select[name="difficulty_2024"]');
        if (d2024) diff = d2024.value;
    }
    if (diff && diff !== defaultDiff) {
        params.set('diff', diff);
    }

    // cart — id@source[:qty], coalescing duplicates
    var cartInputs = form.querySelectorAll('input[name="monsters[]"]');
    if (cartInputs.length) {
        var order = [];
        var counts = {};
        cartInputs.forEach(function(input) {
            var v = (input.value || '').trim();
            if (!v.includes('@')) return;
            if (counts[v] === undefined) {
                counts[v] = 0;
                order.push(v);
            }
            counts[v] += 1;
        });
        if (order.length) {
            var parts = order.map(function(key) {
                return counts[key] > 1 ? key + ':' + counts[key] : key;
            });
            params.set('cart', parts.join(','));
        }
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

// --- Copy link --------------------------------------------------------------

// Delegated handler because the result panel is swapped in by HTMX and the
// button lives inside it. Falls back to a no-op when the clipboard API is
// unavailable (insecure context, ancient browser).
function initCopyLink() {
    document.body.addEventListener('click', function(evt) {
        var btn = evt.target.closest('#copy-link-btn');
        if (!btn) return;
        evt.preventDefault();
        copyCurrentURL(btn);
    });
}

function copyCurrentURL(btn) {
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

    // Ruleset toggle
    var rulesetRadios = document.querySelectorAll('input[name="ruleset"]');
    var difficulty2024Panel = document.getElementById('difficulty-2024-panel');
    var difficulty2014Panel = document.getElementById('difficulty-2014-panel');

    function updateRulesetUI() {
        var checked = document.querySelector('input[name="ruleset"]:checked');
        if (!checked) return;
        var is2024 = checked.value === '2024';
        if (difficulty2024Panel) difficulty2024Panel.classList.toggle('hidden', !is2024);
        if (difficulty2014Panel) difficulty2014Panel.classList.toggle('hidden', is2024);
    }

    rulesetRadios.forEach(function(radio) {
        radio.addEventListener('change', updateRulesetUI);
    });

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
                div.innerHTML = '<label>Personaggio ' + (count + 1) + '</label>' +
                    '<input type="number" name="character_levels" value="3" min="1" max="20" required/>';
                charContainer.appendChild(div);
                updateCharacterControls();
            }
        });

        if (removeCharBtn) {
            removeCharBtn.addEventListener('click', function() {
                if (charContainer.children.length > 1) {
                    charContainer.removeChild(charContainer.lastElementChild);
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
