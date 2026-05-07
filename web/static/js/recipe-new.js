sectionContainer = document.getElementById("ingredients-sections-container")

const inputElement = document.querySelector("[data-input-image]");
const imgEl = document.querySelector("[data-image]");
const btnRemove = document.querySelector("[data-button-image-remove]")

if (window.recipeContext.isEditMode) {
    // TODO
} else {
    if (window.recipeContext.serverData) {
        populateFormWithRecipe(window.recipeContext.serverData);
    } else {
        populateNewForm();
    }
}

setupFormSubmitHandler();
setupPhotoUploadListener();

function addIngredient(ingredientsContainer, isInit = false, value = null) {
    const template = document.getElementById("tmpl-ingredient");
    const clone = document.importNode(template.content, true);
    const removeButton = clone.querySelector("button");
    const ingredient = clone.querySelector("[data-ingredient]");
    removeButton.addEventListener("click", function() {
        ingredient.remove();
    });

    const ingredientInput = ingredient.querySelector("[data-ingredient-input]");
    const amountInput = ingredient.querySelector("[data-amount-input]");

    if (value != null) {
        ingredientInput.value = value.Name;
        amountInput.value = value.Amount;
    }

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

function addSection(isInit = false, sectionData = null) {
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
    const firstIngredient = addIngredient(
        ingredientsContainer,
        isInit,
        sectionData && sectionData.Ingredients.length > 0 ? sectionData.Ingredients[0] : null
    );
    const firstIngredientInput = firstIngredient.querySelector("[data-ingredient-input]");
    container.appendChild(clone);
    const sectionHeadingInput = section.querySelector("[data-section-heading-input]");
    if (sectionData) {
        sectionHeadingInput.value = sectionData.Heading;
        for (i = 1; i < sectionData.Ingredients.length; i++) {
            addIngredient(ingredientsContainer, isInit, sectionData.Ingredients[i]);
        }
    }
    if (!isInit) {
        sectionHeadingInput.focus();
    }
    sectionHeadingInput.addEventListener('keydown', e => {
        if (e.key !== 'Enter') return;
        e.preventDefault();
        firstIngredientInput.focus();
    });
}


function populateFormWithRecipe(recipe) {
    document.querySelector("[name=title]").value = recipe.Title;
    if (recipe.MealTypes) {
        document.querySelectorAll("[name='meal-types[]']").forEach(inputElement => {
            if (recipe.MealTypes.includes(inputElement.value)) {
                inputElement.checked = true;
            }
        })
    }
    if (recipe.DietaryTags) {
        document.querySelectorAll("[name='diets[]']").forEach(inputElement => {
            if (recipe.DietaryTags.includes(inputElement.value)) {
                inputElement.checked = true;
            }
        })
    }
    if (recipe.OtherTags) {
        document.querySelectorAll("[name='tags[]']").forEach(inputElement => {
            if (recipe.OtherTags.includes(inputElement.value)) {
                inputElement.checked = true;
            }
        })
    }
    document.querySelector("[name=cook-time]").value = recipe.CookTimeSeconds;
    document.querySelector("[name='prep-time']").value = recipe.PrepTimeSeconds / 3600;
    document.querySelector("[name=description]").value = recipe.Description;
    document.querySelector("[name=servings]").value = recipe.Servings;
    recipe.IngredientSections.forEach(section => {
        addSection(true, section);
    });
    document.querySelector("[name=instructions]").value = recipe.Instructions;
}

function populateNewForm() {
    addSection(true);
}

function setupFormSubmitHandler() {
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
        ingredientsInput.value = JSON.stringify(sections);
    });
}

function setupPhotoUploadListener() {
    const showPreview = (src) => {
        imgEl.src = src;
        imgEl.style.display = "block";
        btnRemove.style.display = "block";
    }

    const hidePreview = () => {
        imgEl.src = "";
        imgEl.style.display = "none";
    }

    const updatePreview = () => {
        URL.revokeObjectURL(imgEl.src);
        const file = inputElement.files && inputElement.files[0]
        if (!file) {
            return;
        }
        showPreview(URL.createObjectURL(file));
        btnRemove.style.display = "block";
    }

    updatePreview();

    inputElement.addEventListener("input", updatePreview);

    btnRemove.addEventListener("click", () => {
        inputElement.value = "";
        btnRemove.style.display = "none";
        const existingURL = window.recipeContext?.imageSrc;
        if (existingURL) {
            showPreview(existingURL);
        } else {
            hidePreview();
        }
    });
}
