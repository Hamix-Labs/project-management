import { parseTaskDraftDetail, parseTaskDraftSummaryList } from "./parseTaskApiDrafts";
import {
  parseTaskComposePayload,
  parseTaskTemplateDetail,
  parseTaskTemplateInstantiateResponse,
  parseTaskTemplateSummaryList,
} from "./parseTaskApiTemplates";

export { parseComposePayloadCore } from "./parseTaskApiCompose";
export { fetchNamedEntityJson, fetchNamedEntityVoid } from "./namedEntityClient";

export type NamedEntityParsers<TSummary, TDetail> = {
  parseSummaryList: (raw: unknown) => TSummary[];
  parseDetail: (raw: unknown) => TDetail;
};

export const taskDraftEntityParsers = {
  parseSummaryList: parseTaskDraftSummaryList,
  parseDetail: parseTaskDraftDetail,
} satisfies NamedEntityParsers<
  ReturnType<typeof parseTaskDraftSummaryList>[number],
  ReturnType<typeof parseTaskDraftDetail>
>;

export const taskTemplateEntityParsers = {
  parseSummaryList: parseTaskTemplateSummaryList,
  parseDetail: parseTaskTemplateDetail,
} satisfies NamedEntityParsers<
  ReturnType<typeof parseTaskTemplateSummaryList>[number],
  ReturnType<typeof parseTaskTemplateDetail>
>;

export { parseTaskComposePayload, parseTaskTemplateInstantiateResponse };
