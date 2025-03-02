function formatDates() {
    document.querySelectorAll(".format-date").forEach((element) => {
        let rawDate = element.dataset.date.trim(); // Get raw date
        let formatType = element.dataset.format;  // Get format type

        if (!rawDate || !formatType) return;

        // ✅ Convert backend date format to a valid ISO format
        let formattedRawDate = rawDate.replace(" ", "T"); // Replace space with 'T' for ISO
        let date = new Date(formattedRawDate); // Parse to Date object

        if (isNaN(date.getTime())) {
            element.innerHTML = "Invalid Date"; // Handle parsing failure
            return;
        }

        // Define different formats
        let formattedDate;
        if (formatType === "long") {
            formattedDate = date.toLocaleDateString('en-US', { day: '2-digit', month: 'long', year: 'numeric' }); // February 27, 2025
        } else if (formatType === "datetime") {
            formattedDate = `${date.getDate()} ${date.toLocaleString('en', { month: 'short' })}, ${date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: true })}`; // 27 Feb, 10:33 AM
        } else if (formatType === "short") {
            formattedDate = `${date.getDate()} ${date.toLocaleString('en', { month: 'short' })}`; // 27 Feb
        } else if (formatType === "custom") {
            formattedDate = `${date.toLocaleString('en', { month: 'short' })} ${date.getDate()}, ${date.getFullYear()}`; // Example: Mar 2, 2025
        }

        element.innerHTML = formattedDate;
    });
}

// Run function when page loads
document.addEventListener("DOMContentLoaded", formatDates);
function showCancelPrompt() {
    document.getElementById('cancelModal').classList.remove('hidden');
    document.body.style.overflow = 'hidden'; // Prevent background scrolling
}