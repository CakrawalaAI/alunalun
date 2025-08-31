import type { Timestamp } from "@bufbuild/protobuf/wkt";
import { timestampDate } from "@bufbuild/protobuf/wkt";

/**
 * Convert a protobuf Timestamp to ISO 8601 string
 * @param timestamp - The protobuf Timestamp object
 * @returns ISO 8601 formatted date string
 */
export function protoTimestampToISO(timestamp?: Timestamp): string {
  if (!timestamp) {
    return new Date().toISOString();
  }

  try {
    return timestampDate(timestamp).toISOString();
  } catch (error) {
    console.error("Failed to convert timestamp:", error);
    return new Date().toISOString();
  }
}

/**
 * Convert a protobuf Timestamp to Date object
 * @param timestamp - The protobuf Timestamp object
 * @returns JavaScript Date object
 */
export function protoTimestampToDate(timestamp?: Timestamp): Date {
  if (!timestamp) {
    return new Date();
  }

  try {
    return timestampDate(timestamp);
  } catch (error) {
    console.error("Failed to convert timestamp:", error);
    return new Date();
  }
}

/**
 * Format a protobuf Timestamp for display
 * @param timestamp - The protobuf Timestamp object
 * @param locale - The locale to use for formatting (default: 'en-US')
 * @returns Formatted date string
 */
export function formatProtoTimestamp(
  timestamp?: Timestamp,
  locale = "en-US",
): string {
  const date = protoTimestampToDate(timestamp);

  return new Intl.DateTimeFormat(locale, {
    year: "numeric",
    month: "long",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
}

/**
 * Get relative time from a protobuf Timestamp (e.g., "2 hours ago")
 * @param timestamp - The protobuf Timestamp object
 * @returns Relative time string
 */
export function getRelativeTime(timestamp?: Timestamp): string {
  if (!timestamp) {
    return "unknown";
  }

  const date = protoTimestampToDate(timestamp);
  const now = new Date();
  const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (diffInSeconds < 60) {
    return "just now";
  } else if (diffInSeconds < 3600) {
    const minutes = Math.floor(diffInSeconds / 60);
    return `${minutes} minute${minutes > 1 ? "s" : ""} ago`;
  } else if (diffInSeconds < 86400) {
    const hours = Math.floor(diffInSeconds / 3600);
    return `${hours} hour${hours > 1 ? "s" : ""} ago`;
  } else if (diffInSeconds < 2592000) {
    const days = Math.floor(diffInSeconds / 86400);
    return `${days} day${days > 1 ? "s" : ""} ago`;
  } else if (diffInSeconds < 31536000) {
    const months = Math.floor(diffInSeconds / 2592000);
    return `${months} month${months > 1 ? "s" : ""} ago`;
  } else {
    const years = Math.floor(diffInSeconds / 31536000);
    return `${years} year${years > 1 ? "s" : ""} ago`;
  }
}
