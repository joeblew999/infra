// XTemplate Project JavaScript
console.log('🚀 XTemplate project loaded!');

// Add any interactive functionality here
document.addEventListener('DOMContentLoaded', function() {
    // Example: highlight current page in navigation
    const currentPath = window.location.pathname;
    const navLinks = document.querySelectorAll('nav a');

    navLinks.forEach(link => {
        if (link.getAttribute('href') === currentPath ||
            (currentPath === '/' && link.getAttribute('href') === '/')) {
            link.style.textDecoration = 'underline';
        }
    });
});
