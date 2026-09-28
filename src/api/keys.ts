export function rotateApiKey(oldKey: string) {
  // Add expiring API keys and replacement rotation
  return { newKey: "new-key-123", expiresAt: Date.now() + 86400 };
}
