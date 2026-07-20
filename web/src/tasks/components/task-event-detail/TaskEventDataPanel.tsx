import { useId, useRef, useState } from "react";
import type { KeyboardEvent } from "react";
import type { TaskEvent } from "@/types/task";
import {
  parseCycleTerminalOverview,
  parsePhaseEventOverview,
} from "../../task-events/parsePhaseEventOverview";
import { CycleTerminalOverviewBody } from "./CycleTerminalOverviewBody";
import { GenericEventDataOverview } from "./GenericEventDataOverview";
import { PhaseEventOverviewBody } from "./PhaseEventOverviewBody";

type TabId = "overview" | "json";
const TAB_ORDER: TabId[] = ["overview", "json"];

function EventDataOverviewSwitch({ event }: { event: TaskEvent }) {
  switch (event.type) {
    case "phase_completed":
    case "phase_failed": {
      const model = parsePhaseEventOverview(event.type, event.data);
      return model ? (
        <PhaseEventOverviewBody model={model} />
      ) : (
        <GenericEventDataOverview data={event.data} />
      );
    }
    case "cycle_failed":
    case "cycle_completed": {
      const model = parseCycleTerminalOverview(event.type, event.data);
      return model ? (
        <CycleTerminalOverviewBody model={model} />
      ) : (
        <GenericEventDataOverview data={event.data} />
      );
    }
    default:
      return <GenericEventDataOverview data={event.data} />;
  }
}

export function TaskEventDataPanel({ event }: { event: TaskEvent }) {
  const dataJson = JSON.stringify(event.data, null, 2);
  const baseId = useId();
  const tabOverviewId = `${baseId}-tab-overview`;
  const tabJsonId = `${baseId}-tab-json`;
  const panelOverviewId = `${baseId}-panel-overview`;
  const panelJsonId = `${baseId}-panel-json`;
  const tabRefs = useRef<Record<TabId, HTMLButtonElement | null>>({
    overview: null,
    json: null,
  });

  const [tab, setTab] = useState<TabId>("overview");

  const selectTab = (nextTab: TabId, focus = false) => {
    setTab(nextTab);
    if (focus) {
      tabRefs.current[nextTab]?.focus();
    }
  };

  const handleTabKeyDown = (eventKey: KeyboardEvent<HTMLButtonElement>) => {
    const currentIndex = TAB_ORDER.indexOf(tab);
    let nextTab: TabId | undefined;

    if (eventKey.key === "ArrowRight") {
      nextTab = TAB_ORDER[(currentIndex + 1) % TAB_ORDER.length];
    } else if (eventKey.key === "ArrowLeft") {
      nextTab =
        TAB_ORDER[(currentIndex - 1 + TAB_ORDER.length) % TAB_ORDER.length];
    } else if (eventKey.key === "Home") {
      nextTab = TAB_ORDER[0];
    } else if (eventKey.key === "End") {
      nextTab = TAB_ORDER[TAB_ORDER.length - 1];
    }

    if (!nextTab) return;
    eventKey.preventDefault();
    selectTab(nextTab, true);
  };

  return (
    <div className="task-event-detail-data-block">
      <h3 className="task-detail-subheading task-event-data-heading">
        <span>Event data</span>
      </h3>

      <div
        className="task-event-data-tabs"
        role="tablist"
        aria-label="Event payload view"
      >
        <button
          type="button"
          id={tabOverviewId}
          role="tab"
          aria-selected={tab === "overview"}
          aria-controls={panelOverviewId}
          tabIndex={tab === "overview" ? 0 : -1}
          ref={(node) => {
            tabRefs.current.overview = node;
          }}
          className="task-event-data-tab"
          data-active={tab === "overview" ? "true" : undefined}
          onClick={() => selectTab("overview")}
          onKeyDown={handleTabKeyDown}
        >
          Overview
        </button>
        <button
          type="button"
          id={tabJsonId}
          role="tab"
          aria-selected={tab === "json"}
          aria-controls={panelJsonId}
          tabIndex={tab === "json" ? 0 : -1}
          ref={(node) => {
            tabRefs.current.json = node;
          }}
          className="task-event-data-tab"
          data-active={tab === "json" ? "true" : undefined}
          onClick={() => selectTab("json")}
          onKeyDown={handleTabKeyDown}
        >
          Raw JSON
        </button>
      </div>
      <div
        id={panelOverviewId}
        role="tabpanel"
        aria-labelledby={tabOverviewId}
        hidden={tab !== "overview"}
        className="task-event-data-panel"
      >
        <EventDataOverviewSwitch event={event} />
      </div>
      <div
        id={panelJsonId}
        role="tabpanel"
        aria-labelledby={tabJsonId}
        hidden={tab !== "json"}
        className="task-event-data-panel"
      >
        <pre className="task-timeline-data task-event-detail-data-pre">
          {dataJson}
        </pre>
      </div>
    </div>
  );
}
