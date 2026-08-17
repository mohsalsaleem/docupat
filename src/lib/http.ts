export async function requestJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...init,
    headers: { 'content-type': 'application/json', ...init?.headers },
  });
  const data = await response.json();
  if (!response.ok) throw new Error(data.error || `Request failed: ${response.status}`);
  return data;
}
