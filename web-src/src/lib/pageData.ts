declare global {
  interface Window {
    __DATA__: unknown;
  }
}

export function getPageData<T>(): T {
  return window.__DATA__ as T;
}
