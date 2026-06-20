const displayNameInput = document.querySelector("[name=display-name]");
const displayNameSubmitButton = document.getElementById("btn-submit-display-name");

const newPasswordInput = document.querySelector("[name=new-password]");
const msgPasswordTooShort = document.querySelector("[data-msg-pass-too-short]");
const inputConfirmNewPassword = document.querySelector("[name=confirm-new-password]");
const msgConfirmPasswordDoesNotMatch = document.querySelector("[data-msg-confirm-password-does-not-match]")

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

const classHideFormError = "form-error--hidden";

function updateMsgConfirmPasswordDoesNotMatch() {
    const doesMatch = newPasswordInput.value == inputConfirmNewPassword.value;
    if (doesMatch) {
        msgConfirmPasswordDoesNotMatch.classList.add(classHideFormError)
    } else {
        msgConfirmPasswordDoesNotMatch.classList.remove(classHideFormError)
    }
}

newPasswordInput.addEventListener("input", function() {
    const isPasswordTooShort = newPasswordInput.value.length < 15;
    if (isPasswordTooShort) {
        msgPasswordTooShort.classList.remove(classHideFormError);
    } else {
        msgPasswordTooShort.classList.add(classHideFormError);
    }
    updateMsgConfirmPasswordDoesNotMatch();
});

inputConfirmNewPassword.addEventListener("input", updateMsgConfirmPasswordDoesNotMatch);
