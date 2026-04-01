const textareas = document.querySelectorAll('textarea');

textareas.forEach(textarea => {
    textarea.addEventListener('input', function() {
        this.style.height = 'auto';
        this.style.height = this.scrollHeight + 2 + 'px';
    });
})

const fileUploads = document.querySelectorAll('input[type="file"]');

fileUploads.forEach(inputEl => {
    inputEl.addEventListener('input', function(event) {
        console.log("detected upload...");
        if (!event.target.files || !event.target.files[0]) {
            return;
        }
        console.log("now updating image...");
        const imgEl = inputEl.parentNode.querySelector('img');
        imgEl.src = URL.createObjectURL(event.target.files[0]);
        imgEl.style.display = "block";
    });
});
