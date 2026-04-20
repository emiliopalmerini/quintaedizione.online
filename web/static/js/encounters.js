// Encounters Calculator JavaScript
// Depends on main.js for: initThemeToggle, initPatreonBanner

document.addEventListener('DOMContentLoaded', function() {
    initEncounterForm();
    initSteppers();
    initAutoCalculate();

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

// Debounced auto-calculate: fires custom 'calculate' event that HTMX listens to
function initAutoCalculate() {
    var form = document.getElementById('encounter-form');
    if (!form || typeof htmx === 'undefined') return;

    var timer = null;
    function scheduleCalculate() {
        clearTimeout(timer);
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
