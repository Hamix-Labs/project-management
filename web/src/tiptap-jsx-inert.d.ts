import "react";

declare module "react" {
  // TipTap chrome sets `inert`; React 18's HTMLAttributes omit it.
  // eslint-disable-next-line @typescript-eslint/no-unused-vars -- matches React's generic
  interface HTMLAttributes<T> {
    inert?: boolean | "" | undefined;
  }
}
