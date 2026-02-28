const idPrefix = "auth-menu-";
const openClass = "open";

const menu = document.getElementById("auth-menu")
const profileBtn = document.getElementById(idPrefix + "profile-button")
const dropdown = document.getElementById(idPrefix + "dropdown")

profileBtn.addEventListener("click", function(e) {
    e.stopPropagation();
    dropdown.classList.toggle(openClass);
});

document.addEventListener("click", function(e) {
    if (!menu.contains(e.target)) {
        dropdown.classList.remove(openClass);
    }
})

document.addEventListener("keydown", function(e) {
    if (e.key == "Escape") {
        dropdown.classList.remove(openClass);
        if (menu.contains(e.target)) {
            profileBtn.focus();
        }
    }
})


