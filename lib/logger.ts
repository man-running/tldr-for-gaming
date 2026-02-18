type LogLevel = "debug" | "info" | "warn" | "error";

interface LogContext {
	[key: string]: unknown;
}

class Logger {
	private isDebugEnabled = process.env.RSS_DEBUG === "true";

	private shouldLog(level: LogLevel): boolean {
		if (level === "debug") {
			return this.isDebugEnabled;
		}
		return true; // Always log info, warn, and error
	}

	private formatMessage(
		level: LogLevel,
		message: string,
		context?: LogContext,
	): string {
		const logEntry = {
			timestamp: new Date().toISOString(),
			level: level.toUpperCase(),
			message,
			...(context && { context }),
		};
		return JSON.stringify(logEntry);
	}

	debug(message: string, context?: LogContext): void {
		if (this.shouldLog("debug")) {
			console.log(this.formatMessage("debug", message, context));
		}
	}

	info(message: string, context?: LogContext): void {
		if (this.shouldLog("info")) {
			console.log(this.formatMessage("info", message, context));
		}
	}

	warn(message: string, context?: LogContext): void {
		if (this.shouldLog("warn")) {
			console.warn(this.formatMessage("warn", message, context));
		}
	}

	error(message: string, error?: Error, context?: LogContext): void {
		if (this.shouldLog("error")) {
			const errorContext = {
				...context,
				errorMessage: error?.message,
				errorStack: error?.stack,
			};
			console.error(this.formatMessage("error", message, errorContext));
		}
	}
}

export const logger = new Logger();
