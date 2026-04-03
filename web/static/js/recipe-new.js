sectionContainer = document.getElementById("ingredients-sections-container")

function addIngredient(ingredientsContainer, isInit = false) {
    const template = document.getElementById("tmpl-ingredient");
    const clone = document.importNode(template.content, true);
    const removeButton = clone.querySelector("button");
    const ingredient = clone.querySelector("[data-ingredient]");
    removeButton.addEventListener("click", function() {
        ingredient.remove();
    });

    const ingredientInput = ingredient.querySelector("[data-ingredient-input]");
    const amountInput = ingredient.querySelector("[data-amount-input]");

    amountInput.addEventListener('keydown', e => {
        if (e.key !== 'Enter') return;
        e.preventDefault();
        addIngredient(ingredientsContainer);
    });
    ingredientInput.addEventListener('keydown', e => {
        if (e.key !== 'Enter') return;
        e.preventDefault();
        addIngredient(ingredientsContainer);
    });

    ingredientsContainer.appendChild(clone);

    if (!isInit) {
        ingredientInput.focus();
    }

    return ingredient;
}

function addSection(isInit = false) {
    const template = document.getElementById("tmpl-ingredients-section");
    const clone = document.importNode(template.content, true);
    const container = document.getElementById("ingredients-sections-container");
    const ingredientsContainer = clone.querySelector("[data-ingredients-container]");
    const removeSectionButton = clone.querySelector("[data-action='remove-section']");
    const newIngredientButton = clone.querySelector("[data-action='new-ingredient']");
    const section = clone.querySelector("[data-section]");
    removeSectionButton.addEventListener("click", function() {
        section.remove();
    });
    newIngredientButton.addEventListener("click", function() {
        addIngredient(ingredientsContainer);
    });
    const firstIngredient = addIngredient(ingredientsContainer, isInit);
    const firstIngredientInput = firstIngredient.querySelector("[data-ingredient-input]");
    container.appendChild(clone);
    const sectionHeadingInput = section.querySelector("[data-section-heading-input]");
    if (!isInit) {
        sectionHeadingInput.focus();
    }
    sectionHeadingInput.addEventListener('keydown', e => {
        if (e.key !== 'Enter') return;
        e.preventDefault();
        firstIngredientInput.focus();
    });
}

addSection(true);

document.querySelector("[data-new-recipe-form]").addEventListener("submit", (e) => {
    // TODO: add client side check to make sure form data is valid

    const sections = Array.from(document.querySelectorAll("[data-section]")).map(section => ({
        heading: section.querySelector("[data-section-heading-input]").value.trim(),
        ingredients: Array.from(section.querySelectorAll("[data-ingredient]")).map(ingredient => ({
            name: ingredient.querySelector("[data-ingredient-input]").value.trim(),
            amount: ingredient.querySelector("[data-amount-input]").value.trim(),
        }))
    }));
    const ingredientsInput = document.querySelector("[data-ingredients-input]");
    console.log(JSON.stringify(sections));
    ingredientsInput.value = JSON.stringify(sections);
});
