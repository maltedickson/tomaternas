const dropdowns = document.querySelectorAll("[data-dropdown]");
console.debug(dropdowns);

const classOpen = "open";

dropdowns.forEach(dropdown => {
    const trigger = dropdown.querySelector("[data-dropdown-trigger]");
    const menu = dropdown.querySelector("[data-dropdown-menu]");

    document.addEventListener("click", function(e) {
        if (trigger.contains(e.target)) {
            e.stopPropagation();
            menu.classList.toggle(classOpen);
        } else {
            menu.classList.remove(classOpen);
        }
    })
    document.addEventListener("keydown", function(e) {
        if (e.key == "Escape") {
            menu.classList.remove(openClass);
            if (dropdowns.contains(e.target)) {
                trigger.focus();
            }
        }
    })
})

