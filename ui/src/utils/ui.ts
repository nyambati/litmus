import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

/**
 * Global API endpoint configuration
 */
export const API = import.meta.env.DEV ? "http://localhost:8080" : "";

/**
 * Tailwind CSS class merging utility
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

/**
 * Formats a timestamp into a human-readable relative time string
 */
export function formatAge(ts: number | null): string {
  if (!ts) return "";
  const secs = Math.floor((Date.now() - ts) / 1000);
  if (secs < 60) return "just now";
  if (secs < 3600) return `${Math.floor(secs / 60)}m ago`;
  if (secs < 86400) return `${Math.floor(secs / 3600)}h ago`;
  return `${Math.floor(secs / 86400)}d ago`;
}

/**
 * Ensures a promise takes at least ms to resolve for smooth UI transitions
 */
export const minDelay = <T>(promise: Promise<T>, ms = 300): Promise<T> => {
  return Promise.all([
    promise,
    new Promise((resolve) => setTimeout(resolve, ms)),
  ]).then(([res]) => res);
};
