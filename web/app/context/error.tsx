import { createContext, useContext, useState } from "react";
import type { ReactNode } from "react";

interface ErrorInfo {
    message: string;
    source?: string | null;
    stack?: string;
    timestamp: string;
}

interface ErrorContextValue {
    errors: ErrorInfo[];
    addError: (err: ErrorInfo) => void;
    clearErrors: () => void;
}

const ErrorContext = createContext<ErrorContextValue | null>(null);

interface ErrorProviderProps {
    children: ReactNode;
}

export function ErrorProvider({ children }: ErrorProviderProps) {
    const [errors, setErrors] = useState<ErrorInfo[]>([]);

    const addError = (err: ErrorInfo) => {
        setErrors(prev => {
            if (prev.some(e => e.message === err.message && e.stack === err.stack)) {
                return prev;
            }
            return [...prev, err]
        });
    };

    const clearErrors = () => setErrors([]);

    return (
        <ErrorContext.Provider value={{ errors, addError, clearErrors }}>
            {children}
        </ErrorContext.Provider>
    );
}

export function useErrors(): ErrorContextValue {
    const context = useContext(ErrorContext);
    if (!context) {
        throw new Error("useErrors must be inside ErrorProvider")
    }
    return context;

}