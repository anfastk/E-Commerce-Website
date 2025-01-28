// toast.js

// Toast configuration
const TOAST_DURATION = 3000; // Duration in milliseconds

class ToastManager {
    constructor() {
        this.toastTimeout = null;
        // Don't create container in constructor
        // Will be created on first use
    }

    createToastContainer() {
        // Check if toast container already exists
        let container = document.getElementById('toast-container');
        
        if (!container) {
            container = document.createElement('div');
            container.id = 'toast-container';
            container.className = 'fixed z-50 top-20 left-[60%] transform -translate-x-1/2';
            document.body.appendChild(container);
        }
        
        return container;
    }

    show(message, type = 'success') {
        // Ensure we have a container
        const toastContainer = this.createToastContainer();

        // Clear any existing timeout
        if (this.toastTimeout) {
            clearTimeout(this.toastTimeout);
        }
        
        // Clear existing toasts
        toastContainer.innerHTML = '';

        // Create new toast element
        const toast = document.createElement('div');
        toast.className = `${type === 'success' ? 'bg-green-500' : 'bg-red-500'} text-white px-6 py-3 rounded shadow-lg transition-opacity duration-300`;
        
        // Create message element
        const messageElement = document.createElement('span');
        messageElement.textContent = message;
        toast.appendChild(messageElement);

        // Add toast to container
        toastContainer.appendChild(toast);

        // Show toast
        requestAnimationFrame(() => {
            toast.style.opacity = '1';
        });

        // Hide toast after duration
        this.toastTimeout = setTimeout(() => {
            toast.style.opacity = '0';
            setTimeout(() => {
                if (toastContainer.contains(toast)) {
                    toastContainer.removeChild(toast);
                }
            }, 300);
        }, TOAST_DURATION);
    }

    success(message) {
        this.show(message, 'success');
    }

    error(message) {
        this.show(message, 'error');
    }
}

// Create global toast instance only after DOM is loaded
let toastInstance = null;

function getToast() {
    if (!toastInstance) {
        toastInstance = new ToastManager();
    }
    return toastInstance;
}

// Export to window object
window.toast = {
    success: (message) => getToast().success(message),
    error: (message) => getToast().error(message)
};