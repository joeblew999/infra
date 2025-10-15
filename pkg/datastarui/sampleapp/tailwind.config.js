const path = require('path');
const forkConfig = require(path.resolve(__dirname, '../../.src/datastarui/fork/datastarui/tailwind.config.js'));

/** @type {import('tailwindcss').Config} */
module.exports = {
  ...forkConfig,
  content: [
    "./pages/**/*.{templ,go}",
    path.resolve(__dirname, "../../.src/datastarui/fork/datastarui/components/**/*.{templ,go}"),
    path.resolve(__dirname, "../../.src/datastarui/fork/datastarui/layouts/**/*.{templ,go}"),
    path.resolve(__dirname, "../../.src/datastarui/fork/datastarui/pages/**/*.{templ,go}"),
  ],
};
