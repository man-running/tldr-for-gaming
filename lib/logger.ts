type LogLevel = "debug" | "info" | "warn" | "error";

interface LogContext {
	[key: string]: unknown;
}

class Logger {
	private isDevelopment = process.env.NODE_ENV === "development";
	private isDebugEnabled = process.env.RSS_DEBUG === "true";

	private shouldLog(level: LogLevel): boolean {
		if (level === "debug") {
			return this.isDevelopment && this.isDebugEnabled;
		}
		return this.isDevelopment || level === "error" || level === "warn";
	}

	private formatMessage(
		level: LogLevel,
		message: string,
		context?: LogContext,
	): string {
		const timestamp = new Date().toISOString();
		const contextStr = context ? ` ${JSON.stringify(context)}` : "";
		return `[${timestamp}] ${level.toUpperCase()}: ${message}${contextStr}`;
	}

	debug(message: string, context?: LogContext): void {
		if (this.shouldLog("debug")) {
			console.log(this.formatMessage("debug", message, context));
		}
	}

	info(message: string, context?: LogContext): void {
		if (this.shouldLog("info")) {
			console.info(this.formatMessage("info", message, context));
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
				error: error?.message,
				stack: error?.stack,
			};
			console.error(this.formatMessage("error", message, errorContext));
		}
	}
}

export const logger = new Logger();
