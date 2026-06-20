const displayNameInput = document.querySelector("[name=display-name]");
const displayNameSubmitButton = document.getElementById("btn-submit-display-name");

function updateDisplayNameSubmitButtonState() {
    if (displayNameInput.value === displayNameInput.defaultValue) {
        displayNameSubmitButton.style.display = "none";
    } else {
        displayNameSubmitButton.style.display = "";
    }
}
updateDisplayNameSubmitButtonState();
displayNameInput.addEventListener("input", updateDisplayNameSubmitButtonState);
displayNameInput.addEventListener("reset", updateDisplayNameSubmitButtonState);
