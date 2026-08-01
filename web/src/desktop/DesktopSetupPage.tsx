import { HamixWordmark } from "@/components/layout/HamixWordmark";
import type { DesktopBridge } from "./bridge";
import { DesktopDatabaseForm } from "./DesktopDatabaseForm";
import "./desktopSetup.css";

type Props = {
  bridge: DesktopBridge;
};

export function DesktopSetupPage({ bridge }: Props) {
  return (
    <div className="desktop-setup" data-testid="desktop-setup-page">
      <header className="desktop-setup-header">
        <HamixWordmark />
        <h1 className="desktop-setup-title">Connect your database</h1>
        <p className="desktop-setup-lead">
          Enter a Postgres connection URL to finish setting up Hamix on this
          machine.
        </p>
      </header>
      <DesktopDatabaseForm bridge={bridge} />
    </div>
  );
}
