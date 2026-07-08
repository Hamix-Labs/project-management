let sseLiveForQueries = false;

export function setSseLiveForQueries(connected: boolean): void {
  sseLiveForQueries = connected;
}

export function isSseLiveForQueries(): boolean {
  return sseLiveForQueries;
}
