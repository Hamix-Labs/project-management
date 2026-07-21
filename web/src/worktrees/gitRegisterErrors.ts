import { ApiError } from "@/api";

/** User-facing copy for register-repository mutation failures. */
export function gitRegisterErrorMessage(err: unknown): string {
  if (err instanceof ApiError && err.code === "duplicate") {
    return "This repository is already registered.";
  }
  if (err instanceof ApiError) {
    return err.message;
  }
  return err instanceof Error ? err.message : "Register failed";
}

export function isDuplicateRegisterError(err: unknown): boolean {
  return err instanceof ApiError && err.code === "duplicate";
}
