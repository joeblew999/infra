# npm


Create an npm Account:
       * Go to the official npm website: https://www.npmjs.com/ 
         (https://www.npmjs.com/)
       * Click on the "Sign Up" or "Join" button.
       * Follow the prompts to create a new account. You'll need a username,
         password, and email address.

   2. Generate an npm Token (after logging in):
      Once you have an account and are logged in on the npm website:
       * Go to your account settings.
       * Look for a section related to "Auth Tokens" or "Access Tokens."
       * You can generate a new token there. When generating, you'll typically
         be asked to choose the type of token (e.g., "Publish" token for
         publishing packages).

      Alternatively, and more commonly for publishing, you can generate and
  store the token directly from your command line after creating your account:
       * Open your terminal or command prompt.
       * Run the command: npm login (or npm adduser).
       * Follow the prompts to enter your npm username, password, and email
         address.
       * npm will then securely store an authentication token on your local
         machine (in your user's .npmrc file), which npm publish will
         automatically use. You don't usually need to manually copy and paste
         the token from the website for local publishing.