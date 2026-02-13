interface MathJaxObject {
	typeset?: () => void;
	typesetPromise?: () => Promise<void>;
	startup?: {
		defaultReady?: () => void;
	};
}

declare global {
	interface Window {
		MathJax?: MathJaxObject;
	}
}

export {};
