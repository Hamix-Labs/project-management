import { useNavigate } from "react-router-dom";
import { RegisterRepositoryFirstPrompt } from "@/components/RegisterRepositoryFirstPrompt";

type Props = {
  open: boolean;
  onClose: () => void;
};

/** Worktrees-owned wrapper: register CTA routes to `/repositories?register=1`. */
export function RepositorySetupPrompt({ open, onClose }: Props) {
  const navigate = useNavigate();

  return (
    <RegisterRepositoryFirstPrompt
      open={open}
      onClose={onClose}
      onRegister={() => {
        onClose();
        navigate("/repositories?register=1");
      }}
    />
  );
}
