// WebAuthn and Auth JavaScript functionality

// Base64URL encode/decode helpers
function base64URLEncode(arrayBuffer) {
  return btoa(String.fromCharCode(...new Uint8Array(arrayBuffer)))
    .replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
}

function base64URLDecode(str) {
  // Handle both base64URL and regular base64 from go-webauthn
  if (!str.includes('-') && !str.includes('_')) {
    // Regular base64 - add padding if needed
    const padding = 4 - (str.length % 4);
    if (padding !== 4) {
      str += '='.repeat(padding);
    }
    return Uint8Array.from(atob(str), c => c.charCodeAt(0));
  } else {
    // Base64URL - convert and add padding
    const padding = 4 - (str.length % 4);
    if (padding !== 4) {
      str += '='.repeat(padding);
    }
    str = str.replace(/-/g, '+').replace(/_/g, '/');
    return Uint8Array.from(atob(str), c => c.charCodeAt(0));
  }
}

// Convert credential creation options for WebAuthn API
function processCreationOptions(options) {
  const publicKey = options.publicKey || options;
  
  return {
    ...publicKey,
    challenge: base64URLDecode(publicKey.challenge),
    user: {
      ...publicKey.user,
      id: base64URLDecode(publicKey.user.id)
    },
    excludeCredentials: publicKey.excludeCredentials?.map(cred => ({
      ...cred,
      id: base64URLDecode(cred.id)
    }))
  };
}

// Convert credential request options for WebAuthn API
function processRequestOptions(options) {
  const publicKey = options.publicKey || options;
  
  return {
    ...publicKey,
    challenge: base64URLDecode(publicKey.challenge),
    allowCredentials: publicKey.allowCredentials?.map(cred => ({
      ...cred,
      id: base64URLDecode(cred.id)
    }))
  };
}

// Format credential creation response
function formatCreationResponse(credential) {
  return {
    id: credential.id,
    rawId: btoa(String.fromCharCode(...new Uint8Array(credential.rawId))),
    type: credential.type,
    response: {
      attestationObject: btoa(String.fromCharCode(...new Uint8Array(credential.response.attestationObject))),
      clientDataJSON: btoa(String.fromCharCode(...new Uint8Array(credential.response.clientDataJSON)))
    }
  };
}

// Format credential request response  
function formatRequestResponse(credential) {
  return {
    id: credential.id,
    rawId: btoa(String.fromCharCode(...new Uint8Array(credential.rawId))),
    type: credential.type,
    response: {
      authenticatorData: btoa(String.fromCharCode(...new Uint8Array(credential.response.authenticatorData))),
      clientDataJSON: btoa(String.fromCharCode(...new Uint8Array(credential.response.clientDataJSON))),
      signature: btoa(String.fromCharCode(...new Uint8Array(credential.response.signature))),
      userHandle: credential.response.userHandle ? btoa(String.fromCharCode(...new Uint8Array(credential.response.userHandle))) : null
    }
  };
}

// WebAuthn registration
async function startRegistration(options) {
  const processedOptions = processCreationOptions(options);
  const credential = await navigator.credentials.create({
    publicKey: processedOptions
  });
  return formatCreationResponse(credential);
}

// WebAuthn authentication
async function startAuthentication(options) {
  const processedOptions = processRequestOptions(options);
  const credential = await navigator.credentials.get({
    publicKey: processedOptions
  });
  return formatRequestResponse(credential);
}

// Conditional UI for passive login
async function startConditionalUI() {
  try {
    if (!window.PublicKeyCredential || !PublicKeyCredential.isConditionalMediationAvailable) {
      console.log('Conditional UI not supported');
      return;
    }

    const available = await PublicKeyCredential.isConditionalMediationAvailable();
    if (!available) {
      console.log('Conditional mediation not available');
      return;
    }

    console.log('Starting conditional UI authentication...');
    document.getElementById('log').textContent = 'Ready for passkey login - click in username field';

    const credential = await navigator.credentials.get({
      publicKey: {
        challenge: new Uint8Array(32),
        userVerification: 'preferred',
        timeout: 300000,
      },
      mediation: 'conditional'
    });

    if (credential) {
      document.getElementById('log').textContent = 'Passkey selected! Processing login...';
      
      const response = await fetch('/login/conditional', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({
          credentialId: credential.id
        })
      });
      
      if (response.ok) {
        document.getElementById('log').textContent = 'Logged in successfully! Welcome back.';
        setTimeout(() => {
          window.location.href = '/dashboard';
        }, 2000);
      }
    }
    
  } catch (error) {
    console.error('Conditional UI error:', error);
    if (error.name !== 'AbortError') {
      document.getElementById('log').textContent = 'Conditional login not available. Use register button to create a passkey.';
    }
  }
}

// Session management
async function checkSessionStatus() {
  try {
    const response = await fetch('/session/status');
    // If logged in, the SSE response will update the UI automatically
    if (!response.ok) {
      // Not logged in, start conditional UI for login
      startConditionalUI();
    }
  } catch (error) {
    // Error checking session, assume not logged in
    startConditionalUI();
  }
}

// Test user creation - DISABLED FOR SECURITY
// Use DataStar register flow instead
function createTestUser() {
  document.getElementById('log').textContent = 'Test user creation disabled for security. Use "Register New Passkey" instead.';
}

// Logout
function logout() {
  fetch('/logout', {method: 'POST'})
    .then(() => window.location.href = '/');
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
  checkSessionStatus();
});