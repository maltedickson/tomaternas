sectionContainer = document.getElementById("ingredients-sections-container")

function addIngredient(ingredientsContainer) {
    const template = document.getElementById("tmpl-ingredient");
    const clone = document.importNode(template.content, true);
    const removeButton = clone.querySelector("button");
    const ingredient = clone.querySelector("[data-ingredient]");
    removeButton.addEventListener("click", function() {
        ingredient.remove();
    });
    ingredientsContainer.appendChild(clone);
}

function addSection() {
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
    addIngredient(ingredientsContainer);
    container.appendChild(clone);
}

addSection()
