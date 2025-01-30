/* ===================================================================SIDE BAR===================================================================== */
function toggleSidebar() {
  const sidebar = document.getElementById("sidebar");
  const navbarItems = document.getElementById("navbar-items");
  const hamburgerMenu = document.getElementById("hamburger-menu");

  // Toggle sidebar visibility
  sidebar.classList.toggle("hidden");

  // Toggle bottom navbar hamburger menu visibility
  hamburgerMenu.classList.toggle("hidden");

  // Toggle main navbar items visibility when sidebar is opened
  navbarItems.classList.toggle("hidden");
}

/* ===================================================================NAV BAR===================================================================== */

function toggleSearchBar() {
  const searchButton = document.getElementById("search-button");
  const searchBar = document.getElementById("search-bar-container");

  // Hide the search button and show the search bar
  searchButton.classList.add("hidden");
  searchBar.classList.remove("hidden");
}

function clearSearch() {
  // Clear the input field
  document.getElementById("search-input").value = "";

  // Hide the search bar and show the search button
  const searchBar = document.getElementById("search-bar-container");
  const searchButton = document.getElementById("search-button");

  searchBar.classList.add("hidden");
  searchButton.classList.remove("hidden");
}
