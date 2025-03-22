function showErrorToast(message) {
    let toast = document.getElementById("errorToast");
    toast.textContent = message;
    toast.classList.remove("hidden");

    setTimeout(() => {
        toast.classList.add("hidden");
    }, 3000);
}

function showToast(message, type) {
    const toastContainer = document.querySelector('.toast-container');
    const toastTemplate = document.getElementById('toast');
    const toast = toastTemplate.cloneNode(true);

    // Set unique ID
    const toastId = 'toast-' + Date.now();
    toast.id = toastId;

    // Set message
    toast.querySelector('.toast-message').textContent = message;

    // Set type class
    toast.className = 'toast';

    if (type === 'error') {
        toast.classList.add('toast-error');
    } else if (type === 'success') {
        toast.classList.add('toast-success');
    }

    // Add to container
    toastContainer.appendChild(toast);

    // Show with slight delay for proper animation
    setTimeout(() => {
        toast.classList.add('show');
    }, 10);

    // Remove after animation completes
    setTimeout(() => {
        toast.classList.remove('show');

        // Remove from DOM after hide animation
        setTimeout(() => {
            toast.remove();
        }, 300);
    }, 3000);
}

function showErrorToast(message) {
    showToast(message, 'error');
}

function showSuccessToast(message) {
    showToast(message, 'success');
}