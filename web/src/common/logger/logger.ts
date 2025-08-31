/**
 * Unified Logger Utility with Error Sanitization
 *
 * Features:
 * - Environment-aware configuration (dev vs prod)
 * - Automatic error sanitization in production
 * - Configurable log levels via environment variables
 * - Sensitive information detection and filtering
 * - User-safe error message generation
 */

type LogLevel = "none" | "error" | "warn" | "info" | "debug";

interface LoggerConfig {
  enabled: boolean;
  level: LogLevel;
  sanitizeErrors: boolean;
}

interface SanitizedError {
  userMessage: string;
  errorCode?: string;
  shouldRetry?: boolean;
}

/**
 * Transport interface for extending logger functionality
 * Allows integration with external services like Sentry, LogRocket, etc.
 */
interface LogTransport {
  log(level: LogLevel, args: any[]): void;
}

class Logger {
  private config: LoggerConfig;
  private readonly logLevels: LogLevel[] = [
    "none",
    "error",
    "warn",
    "info",
    "debug",
  ];
  private transports: LogTransport[] = [];

  constructor() {
    // Check if we're in development or production
    const isDevelopment = import.meta.env.DEV;

    // Configure based on environment variables
    // Note: These values are baked in at build time for production
    this.config = {
      // Logging is enabled by default unless explicitly disabled
      enabled: import.meta.env.VITE_ENABLE_LOGGING !== "false",

      // Default log level: 'debug' in dev, 'error' in prod
      level:
        (import.meta.env.VITE_LOG_LEVEL as LogLevel) ||
        (isDevelopment ? "debug" : "error"),

      // Sanitize errors in production by default
      sanitizeErrors:
        import.meta.env.VITE_SANITIZE_ERRORS === "true" ||
        (!isDevelopment && import.meta.env.VITE_SANITIZE_ERRORS !== "false"),
    };
  }

  /**
   * Add a transport for external logging services
   * Example: logger.addTransport(new SentryTransport())
   */
  addTransport(transport: LogTransport): void {
    this.transports.push(transport);
  }

  /**
   * Remove a transport
   */
  removeTransport(transport: LogTransport): void {
    const index = this.transports.indexOf(transport);
    if (index > -1) {
      this.transports.splice(index, 1);
    }
  }

  /**
   * Determines if a log should be output based on current level
   */
  private shouldLog(level: LogLevel): boolean {
    if (!this.config.enabled || this.config.level === "none") {
      return false;
    }

    const currentLevelIndex = this.logLevels.indexOf(this.config.level);
    const requestedLevelIndex = this.logLevels.indexOf(level);

    return requestedLevelIndex <= currentLevelIndex && requestedLevelIndex > 0;
  }

  /**
   * Sanitizes arguments to remove sensitive information
   */
  private sanitizeArgs(...args: any[]): any[] {
    if (!this.config.sanitizeErrors) {
      return args;
    }

    return args.map((arg) => {
      if (typeof arg === "string") {
        return this.sanitizeString(arg);
      }
      if (arg instanceof Error) {
        return this.sanitizeErrorObject(arg);
      }
      if (typeof arg === "object" && arg !== null) {
        return this.sanitizeObject(arg);
      }
      return arg;
    });
  }

  /**
   * Sanitizes a string to remove sensitive patterns
   */
  private sanitizeString(str: string): string {
    if (!this.containsSensitiveInfo(str)) {
      return str;
    }

    let sanitized = str;

    // Replace sensitive patterns
    sanitized = sanitized.replace(/\/api\/[^\s]*/gi, "/api/***");
    sanitized = sanitized.replace(/localhost:\d+/gi, "localhost:***");
    sanitized = sanitized.replace(
      /\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/g,
      "***.***.***.***",
    );
    sanitized = sanitized.replace(/Bearer\s+[^\s]+/gi, "Bearer ***");
    sanitized = sanitized.replace(/([a-zA-Z0-9_-]{20,})/g, "***"); // Long tokens

    return sanitized;
  }

  /**
   * Sanitizes an Error object
   */
  private sanitizeErrorObject(error: Error): object {
    return {
      name: error.name,
      message: this.sanitizeString(error.message),
      // Never include stack trace in production
      ...(this.config.sanitizeErrors ? {} : { stack: error.stack }),
    };
  }

  /**
   * Recursively sanitizes an object
   */
  private sanitizeObject(obj: any, depth = 0): any {
    if (depth > 5) {
      return "[Object too deep]";
    }

    const sanitized: any = {};

    for (const key in obj) {
      if (Object.hasOwn(obj, key)) {
        // Skip sensitive keys entirely
        if (/password|secret|token|key|auth/i.test(key)) {
          sanitized[key] = "***";
        } else if (typeof obj[key] === "string") {
          sanitized[key] = this.sanitizeString(obj[key]);
        } else if (typeof obj[key] === "object" && obj[key] !== null) {
          sanitized[key] = this.sanitizeObject(obj[key], depth + 1);
        } else {
          sanitized[key] = obj[key];
        }
      }
    }

    return sanitized;
  }

  /**
   * Checks if a string contains sensitive information
   */
  private containsSensitiveInfo(message: string): boolean {
    const sensitivePatterns = [
      /\/api\//i, // API endpoints
      /localhost/i, // Local URLs
      /\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/, // IP addresses
      /port\s*:\s*\d+/i, // Port numbers
      /Bearer\s+/i, // Auth tokens
      /password/i, // Passwords
      /secret/i, // Secrets
      /token/i, // Tokens
      /key/i, // Keys
      /postgres/i, // Database names
      /\.go:\d+/, // Go stack traces
      /\.ts:\d+/, // TypeScript stack traces
      /at\s+\w+\s+\(/, // Stack trace patterns
    ];

    return sensitivePatterns.some((pattern) => pattern.test(message));
  }

  /**
   * Send log to all registered transports
   */
  private sendToTransports(level: LogLevel, args: any[]): void {
    for (const transport of this.transports) {
      try {
        transport.log(level, args);
      } catch (error) {
        // Prevent transport errors from breaking the app
        console.error("Transport error:", error);
      }
    }
  }

  /**
   * Debug level logging (most verbose)
   */
  debug(...args: any[]): void {
    if (this.shouldLog("debug")) {
      const sanitizedArgs = this.sanitizeArgs(...args);
      console.log("[DEBUG]", ...sanitizedArgs);
      this.sendToTransports("debug", sanitizedArgs);
    }
  }

  /**
   * Info level logging
   */
  info(...args: any[]): void {
    if (this.shouldLog("info")) {
      const sanitizedArgs = this.sanitizeArgs(...args);
      console.info("[INFO]", ...sanitizedArgs);
      this.sendToTransports("info", sanitizedArgs);
    }
  }

  /**
   * Warning level logging
   */
  warn(...args: any[]): void {
    if (this.shouldLog("warn")) {
      const sanitizedArgs = this.sanitizeArgs(...args);
      console.warn("[WARN]", ...sanitizedArgs);
      this.sendToTransports("warn", sanitizedArgs);
    }
  }

  /**
   * Error level logging
   */
  error(...args: any[]): void {
    if (this.shouldLog("error")) {
      const sanitizedArgs = this.sanitizeArgs(...args);
      console.error("[ERROR]", ...sanitizedArgs);
      this.sendToTransports("error", sanitizedArgs);
    }
  }

  /**
   * Sanitizes an error and returns a user-safe message
   * This is the main method for generating user-facing error messages
   */
  sanitizeError(error: unknown): SanitizedError {
    // Log the full error internally (will be sanitized if in production)
    this.error("Error occurred:", error);

    // Handle different error types and return user-safe messages
    if (error instanceof Error) {
      const message = error.message.toLowerCase();

      // Network/connectivity errors
      if (message.includes("network") || message.includes("fetch failed")) {
        return {
          userMessage:
            "Connection error. Please check your internet connection.",
          errorCode: "NETWORK_ERROR",
          shouldRetry: true,
        };
      }

      // Timeout errors
      if (message.includes("timeout")) {
        return {
          userMessage: "Request timed out. Please try again.",
          errorCode: "TIMEOUT",
          shouldRetry: true,
        };
      }

      // Authentication errors
      if (message.includes("unauthorized") || message.includes("401")) {
        return {
          userMessage: "Your session has expired. Please log in again.",
          errorCode: "AUTH_ERROR",
          shouldRetry: false,
        };
      }

      // Insufficient resources
      if (message.includes("insufficient") || message.includes("balance")) {
        return {
          userMessage: "Insufficient resources to complete this action.",
          errorCode: "INSUFFICIENT_RESOURCES",
          shouldRetry: false,
        };
      }

      // File/upload errors
      if (message.includes("upload") || message.includes("file")) {
        return {
          userMessage: "File upload failed. Please try again.",
          errorCode: "UPLOAD_ERROR",
          shouldRetry: true,
        };
      }

      // Server errors
      if (message.includes("500") || message.includes("internal server")) {
        return {
          userMessage: "Server error. Please try again later.",
          errorCode: "SERVER_ERROR",
          shouldRetry: true,
        };
      }

      // Rate limiting
      if (message.includes("429") || message.includes("rate limit")) {
        return {
          userMessage: "Too many requests. Please wait a moment and try again.",
          errorCode: "RATE_LIMIT",
          shouldRetry: true,
        };
      }

      // Not found errors
      if (message.includes("404") || message.includes("not found")) {
        return {
          userMessage: "The requested resource was not found.",
          errorCode: "NOT_FOUND",
          shouldRetry: false,
        };
      }

      // Validation errors
      if (message.includes("validation") || message.includes("invalid")) {
        return {
          userMessage: "Please check your input and try again.",
          errorCode: "VALIDATION_ERROR",
          shouldRetry: false,
        };
      }
    }

    // Generic fallback - never expose raw error messages
    return {
      userMessage: "An unexpected error occurred. Please try again.",
      errorCode: "UNKNOWN_ERROR",
      shouldRetry: true,
    };
  }

  /**
   * Gets the current configuration (useful for debugging)
   */
  getConfig(): Readonly<LoggerConfig> {
    return { ...this.config };
  }

  /**
   * Temporarily changes log level (useful for debugging specific issues)
   */
  setLogLevel(level: LogLevel): void {
    if (this.logLevels.includes(level)) {
      this.config.level = level;
    }
  }

  /**
   * Validates if a message contains sensitive info (for development use)
   */
  validateMessage(message: string): { safe: boolean; reason?: string } {
    if (!this.containsSensitiveInfo(message)) {
      return { safe: true };
    }

    // Find which pattern matched
    const patterns = [
      { pattern: /\/api\//i, reason: "Contains API endpoint" },
      { pattern: /localhost/i, reason: "Contains localhost reference" },
      {
        pattern: /\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/,
        reason: "Contains IP address",
      },
      { pattern: /Bearer\s+/i, reason: "Contains auth token" },
      { pattern: /password/i, reason: "Contains password reference" },
      { pattern: /secret/i, reason: "Contains secret reference" },
    ];

    for (const { pattern, reason } of patterns) {
      if (pattern.test(message)) {
        return { safe: false, reason };
      }
    }

    return { safe: false, reason: "Contains sensitive information" };
  }
}

// Export singleton instance
export const logger = new Logger();

// Export types and interfaces for use in other modules
export type { LogLevel, SanitizedError, LogTransport };
