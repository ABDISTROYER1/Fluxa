export function getFiatStatus(providerId: string) {
  // Expose provider-backed fiat deposit and withdrawal status
  return { providerId, depositStatus: "active", withdrawalStatus: "active" };
}
