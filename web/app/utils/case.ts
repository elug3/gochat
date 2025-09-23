

export function snakeToCamel(s: string): string {
    return s.replace(/_([a-z])/g, (_, c) => c.toUpperCase());
}

export function camelToSnake(s: string): string {
    return s.replace(/([A-Z])/g, '_$1').toLowerCase();
}

export function keysToCamel<T extends object>(obj: any): T {
    if (Array.isArray(obj)) {
        return obj.map(v => keysToCamel(v)) as any;
    } else if (obj !== null && obj.constructor === Object) {
        return Object.fromEntries(Object.entries(obj).map(([k, v]) => [snakeToCamel(k), keysToCamel(v)])) as T;
    }
    return obj;
}
  