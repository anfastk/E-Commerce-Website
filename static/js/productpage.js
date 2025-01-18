document.querySelectorAll('.menuButton').forEach((button) => {
  const menuOptions = button.nextElementSibling;

  button.addEventListener('click', (event) => {
    event.stopPropagation();
    menuOptions.classList.toggle('hidden');

    document.querySelectorAll('.menuOptions').forEach((menu) => {
      if (menu !== menuOptions) {
        menu.classList.add('hidden');
      }
    });
  });
});

document.addEventListener('click', () => {
  document.querySelectorAll('.menuOptions').forEach((menu) => {
    menu.classList.add('hidden');
  });
});

function toggleFilterPopup() {
  document.getElementById('filter-popup').classList.toggle('hidden');
}

function toggleMobileMenu() {
  document.getElementById('mobile-menu').classList.toggle('hidden');
}
