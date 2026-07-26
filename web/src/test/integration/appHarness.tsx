import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render } from "@testing-library/react";
import { BrowserRouter, MemoryRouter, Route, Routes } from "react-router-dom";
import { ROUTER_FUTURE_FLAGS } from "@/lib/routerFutureFlags";
import App from "@/app/App";
import {
  TaskCreateModalsLayer,
  TaskDraftsPage,
  TaskHome,
} from "@/tasks";
import { TasksAppProvider } from "@/tasks/app/TasksAppProvider";
import { useTasksApp } from "@/tasks/hooks/useTasksApp";
import { ModalStackProvider } from "@/shared/ModalStackContext";
import { bootstrapUnavailable } from "@/test/handlers/bootstrap";
import { stubEventSource } from "@/test/browserMocks";
import { draftsListEmpty } from "@/test/handlers/drafts";
import { globalGitApiHandlers } from "@/test/handlers/gitMsw";
import { projectsListEmpty } from "@/test/handlers/projects";
import { repoNotConfigured } from "@/test/handlers/repo";
import { appSettingsOk, listCursorModelsOk } from "@/test/handlers/settings";
import { taskStatsEmpty, tasksListEmpty } from "@/test/handlers/tasks";

export function appDefaultHandlers() {
  return [
    bootstrapUnavailable(),
    appSettingsOk(),
    tasksListEmpty(),
    taskStatsEmpty(),
    repoNotConfigured(),
    draftsListEmpty(),
    projectsListEmpty(),
    ...globalGitApiHandlers(),
    listCursorModelsOk(),
  ];
}

export function setupAppTest() {
  stubEventSource();
  try {
    window.sessionStorage.removeItem("hamix_ui_test_mode");
  } catch {
    /* private mode */
  }
}

export function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });
}

export function renderApp() {
  const queryClient = createTestQueryClient();
  return render(
    <QueryClientProvider client={queryClient}>
      <BrowserRouter future={ROUTER_FUTURE_FLAGS}>
        <App />
      </BrowserRouter>
    </QueryClientProvider>,
  );
}

export function renderAppAt(initialEntries: string[]) {
  const queryClient = createTestQueryClient();
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter future={ROUTER_FUTURE_FLAGS} initialEntries={initialEntries}>
        <App />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

function TasksShellRoutes() {
  const app = useTasksApp({ sseLive: false, dataEnabled: true });
  return (
    <TasksAppProvider value={app}>
      <ModalStackProvider>
        <Routes>
          <Route path="/" element={<TaskHome />} />
          <Route path="/drafts" element={<TaskDraftsPage />} />
        </Routes>
        <TaskCreateModalsLayer />
      </ModalStackProvider>
    </TasksAppProvider>
  );
}

function renderTasksShell(initialEntries: string[]) {
  const queryClient = createTestQueryClient();
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter future={ROUTER_FUTURE_FLAGS} initialEntries={initialEntries}>
        <TasksShellRoutes />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

export function renderTasksHome() {
  return renderTasksShell(["/"]);
}

export function renderTasksAt(initialEntries: string[]) {
  return renderTasksShell(initialEntries);
}
