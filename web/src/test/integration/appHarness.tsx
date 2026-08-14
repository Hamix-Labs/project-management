import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render } from "@testing-library/react";
import { Suspense } from "react";
import { BrowserRouter, MemoryRouter, Route, Routes } from "react-router-dom";
import { ROUTER_FUTURE_FLAGS } from "@/lib/routerFutureFlags";
import App from "@/app/App";
import { TaskDraftsPage, TaskHome } from "@/tasks";
import { TasksAppProvider } from "@/tasks/app/TasksAppProvider";
import { useTasksApp } from "@/tasks/hooks/useTasksApp";
import {
  TaskComposeEditPage,
  TaskComposeNewPage,
  TemplateComposeEditPage,
  TemplateComposeNewPage,
} from "@/tasks/pages/TaskComposePage";
import { ModalStackProvider } from "@/shared/ModalStackContext";
import { bootstrapUnavailable } from "@/test/handlers/bootstrap";
import { stubEventSource } from "@/test/browserMocks";
import { draftCreateOk, draftsListEmpty } from "@/test/handlers/drafts";
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
    draftCreateOk({
      id: "draft-auto",
      name: "Autosaved draft",
      created_at: "2026-04-07T10:00:00Z",
      updated_at: "2026-04-07T10:00:00Z",
    }),
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
        <Suspense fallback={null}>
          <Routes>
            <Route path="/" element={<TaskHome />} />
            <Route path="/drafts" element={<TaskDraftsPage />} />
            <Route path="/tasks/new" element={<TaskComposeNewPage />} />
            <Route path="/tasks/:taskId/edit" element={<TaskComposeEditPage />} />
            <Route path="/templates/new" element={<TemplateComposeNewPage />} />
            <Route
              path="/templates/:templateId/edit"
              element={<TemplateComposeEditPage />}
            />
          </Routes>
        </Suspense>
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
