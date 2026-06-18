const dialogEditReview = document.getElementById("dialog-edit-review");
const reviewEditForm = document.querySelector(".review-form");

function showReviewEditForm() {
    reviewEditForm.reset();
    dialogEditReview.showModal();
}

function cancelReviewEdit() {
    dialogEditReview.close();
    reviewEditForm.reset();
}
