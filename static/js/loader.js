// Function to show the loader
function showLoader() {
    document.querySelector('.loader-container').classList.add('active');
  }
  
  // Function to hide the loader
  function hideLoader() {
    document.querySelector('.loader-container').classList.remove('active');
  }
  
  // Example usage with promise/async operations
  async function fetchData() {
    showLoader();
    try {
        // Simulate API call or other async operation
        await new Promise(resolve => setTimeout(resolve, 2000));
        // Your actual data fetching code here
    } catch (error) {
        console.error('Error:', error);
    } finally {
        hideLoader();
    }
  }