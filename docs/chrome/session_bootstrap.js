// Copy this into a Chrome MCP navigate_page initScript after replacing placeholders.
//
// Example placeholders:
//   __SESSION_TOKEN__  -> the player's session token
//   __DISPLAY_NAME__   -> the player's display name
//
// Usage pattern:
// 1. open a fresh about:blank page in a unique isolatedContext
// 2. navigate to /play/<gameId>, /forensic/<gameId>, or /observe/<gameId>
// 3. pass this file's contents as initScript after filling in the placeholders

(() => {
  document.cookie = 'treachery_session=__SESSION_TOKEN__; path=/; SameSite=Lax';
  localStorage.setItem('treachery.session.displayName', '__DISPLAY_NAME__');
})();
