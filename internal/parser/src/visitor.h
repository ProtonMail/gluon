#ifndef PARSER_VISITOR_H
#define PARSER_VISITOR_H

#include <antlr4-runtime.h>
#include <map>
#include "IMAPParserBaseVisitor.h"
#include "imap.pb.h"

namespace parser {

class Visitor : public imap::IMAPParserBaseVisitor {
 public:
  Visitor(const std::map<std::string, std::string>&);

  // 2.2.1. Client Protocol Sender and Server Protocol Receiver
  antlrcpp::Any visitTag(imap::IMAPParser::TagContext*) override;

  // 4.3. String
  antlrcpp::Any visitStringQtd(imap::IMAPParser::StringQtdContext*) override;
  antlrcpp::Any visitStringLit(imap::IMAPParser::StringLitContext*) override;
  antlrcpp::Any visitRawQChar(imap::IMAPParser::RawQCharContext*) override;
  antlrcpp::Any visitEscQChar(imap::IMAPParser::EscQCharContext*) override;
  antlrcpp::Any visitLiteral(imap::IMAPParser::LiteralContext*) override;
  antlrcpp::Any visitAstringRaw(imap::IMAPParser::AstringRawContext*) override;
  antlrcpp::Any visitAstringStr(imap::IMAPParser::AstringStrContext*) override;
  antlrcpp::Any visitNstring(imap::IMAPParser::NstringContext*) override;

  // 6. Client Commands
  antlrcpp::Any visitCommand(imap::IMAPParser::CommandContext*) override;

  // 6.1. Client Commands - Any State
  antlrcpp::Any visitCapability(imap::IMAPParser::CapabilityContext*) override;
  antlrcpp::Any visitNoop(imap::IMAPParser::NoopContext*) override;
  antlrcpp::Any visitLogout(imap::IMAPParser::LogoutContext*) override;

  // 6.2. Client Commands - Not Authenticated State
  antlrcpp::Any visitStartTLS(imap::IMAPParser::StartTLSContext*) override;
  antlrcpp::Any visitAuth(imap::IMAPParser::AuthContext*) override;
  antlrcpp::Any visitLogin(imap::IMAPParser::LoginContext*) override;

  // 6.3. Client Commands - Authenticated State
  antlrcpp::Any visitMboxInbox(imap::IMAPParser::MboxInboxContext*) override;
  antlrcpp::Any visitMboxOther(imap::IMAPParser::MboxOtherContext*) override;
  antlrcpp::Any visitListMboxRaw(imap::IMAPParser::ListMboxRawContext*) override;
  antlrcpp::Any visitListMboxStr(imap::IMAPParser::ListMboxStrContext*) override;
  antlrcpp::Any visitSelect(imap::IMAPParser::SelectContext*) override;
  antlrcpp::Any visitExamine(imap::IMAPParser::ExamineContext*) override;
  antlrcpp::Any visitCreate(imap::IMAPParser::CreateContext*) override;
  antlrcpp::Any visitDel(imap::IMAPParser::DelContext*) override;
  antlrcpp::Any visitRename(imap::IMAPParser::RenameContext*) override;
  antlrcpp::Any visitSub(imap::IMAPParser::SubContext*) override;
  antlrcpp::Any visitUnsub(imap::IMAPParser::UnsubContext*) override;
  antlrcpp::Any visitList(imap::IMAPParser::ListContext*) override;
  antlrcpp::Any visitLsub(imap::IMAPParser::LsubContext*) override;
  antlrcpp::Any visitStatus(imap::IMAPParser::StatusContext*) override;
  antlrcpp::Any visitAppend(imap::IMAPParser::AppendContext*) override;

  // RFC2177 Idle Command
  antlrcpp::Any visitIdle(imap::IMAPParser::IdleContext*) override;
  antlrcpp::Any visitDone(imap::IMAPParser::DoneContext*) override;

  // RFC2971 ID Command
  antlrcpp::Any visitId(imap::IMAPParser::IdContext*) override;
  antlrcpp::Any visitId_param_list(imap::IMAPParser::Id_param_listContext*) override;
  antlrcpp::Any visitId_param_key_pair(imap::IMAPParser::Id_param_key_pairContext* ctx) override;
  antlrcpp::Any visitId_params(imap::IMAPParser::Id_paramsContext* ctx) override;
  antlrcpp::Any visitId_param_key_value(imap::IMAPParser::Id_param_key_valueContext* ctx) override;

  antlrcpp::Any visitFlagList(imap::IMAPParser::FlagListContext*) override;
  antlrcpp::Any visitDateTime(imap::IMAPParser::DateTimeContext*) override;
  antlrcpp::Any visitTime(imap::IMAPParser::TimeContext*) override;
  antlrcpp::Any visitZone(imap::IMAPParser::ZoneContext*) override;
  antlrcpp::Any visitSignPlus(imap::IMAPParser::SignPlusContext*) override;
  antlrcpp::Any visitSignMinus(imap::IMAPParser::SignMinusContext*) override;

  // 6.4. Client Commands - Selected State
  antlrcpp::Any visitCheck(imap::IMAPParser::CheckContext*) override;
  antlrcpp::Any visitClose(imap::IMAPParser::CloseContext*) override;
  antlrcpp::Any visitExpunge(imap::IMAPParser::ExpungeContext*) override;
  antlrcpp::Any visitUidExpunge(imap::IMAPParser::UidExpungeContext*) override;
  antlrcpp::Any visitUnselect(imap::IMAPParser::UnselectContext*) override;
  antlrcpp::Any visitSearch(imap::IMAPParser::SearchContext*) override;
  antlrcpp::Any visitFetch(imap::IMAPParser::FetchContext*) override;
  antlrcpp::Any visitStore(imap::IMAPParser::StoreContext*) override;
  antlrcpp::Any visitCopy(imap::IMAPParser::CopyContext*) override;
  antlrcpp::Any visitMove(imap::IMAPParser::MoveContext*) override;
  antlrcpp::Any visitUid(imap::IMAPParser::UidContext*) override;

  antlrcpp::Any visitSearchKeyAll(imap::IMAPParser::SearchKeyAllContext*) override;
  antlrcpp::Any visitSearchKeyAnswered(imap::IMAPParser::SearchKeyAnsweredContext*) override;
  antlrcpp::Any visitSearchKeyBcc(imap::IMAPParser::SearchKeyBccContext*) override;
  antlrcpp::Any visitSearchKeyBefore(imap::IMAPParser::SearchKeyBeforeContext*) override;
  antlrcpp::Any visitSearchKeyBody(imap::IMAPParser::SearchKeyBodyContext*) override;
  antlrcpp::Any visitSearchKeyCc(imap::IMAPParser::SearchKeyCcContext*) override;
  antlrcpp::Any visitSearchKeyDeleted(imap::IMAPParser::SearchKeyDeletedContext*) override;
  antlrcpp::Any visitSearchKeyFlagged(imap::IMAPParser::SearchKeyFlaggedContext*) override;
  antlrcpp::Any visitSearchKeyFrom(imap::IMAPParser::SearchKeyFromContext*) override;
  antlrcpp::Any visitSearchKeyKeyword(imap::IMAPParser::SearchKeyKeywordContext*) override;
  antlrcpp::Any visitSearchKeyNew(imap::IMAPParser::SearchKeyNewContext*) override;
  antlrcpp::Any visitSearchKeyOld(imap::IMAPParser::SearchKeyOldContext*) override;
  antlrcpp::Any visitSearchKeyOn(imap::IMAPParser::SearchKeyOnContext*) override;
  antlrcpp::Any visitSearchKeyRecent(imap::IMAPParser::SearchKeyRecentContext*) override;
  antlrcpp::Any visitSearchKeySeen(imap::IMAPParser::SearchKeySeenContext*) override;
  antlrcpp::Any visitSearchKeySince(imap::IMAPParser::SearchKeySinceContext*) override;
  antlrcpp::Any visitSearchKeySubject(imap::IMAPParser::SearchKeySubjectContext*) override;
  antlrcpp::Any visitSearchKeyText(imap::IMAPParser::SearchKeyTextContext*) override;
  antlrcpp::Any visitSearchKeyTo(imap::IMAPParser::SearchKeyToContext*) override;
  antlrcpp::Any visitSearchKeyUnanswered(imap::IMAPParser::SearchKeyUnansweredContext*) override;
  antlrcpp::Any visitSearchKeyUndeleted(imap::IMAPParser::SearchKeyUndeletedContext*) override;
  antlrcpp::Any visitSearchKeyUnflagged(imap::IMAPParser::SearchKeyUnflaggedContext*) override;
  antlrcpp::Any visitSearchKeyUnkeyword(imap::IMAPParser::SearchKeyUnkeywordContext*) override;
  antlrcpp::Any visitSearchKeyUnseen(imap::IMAPParser::SearchKeyUnseenContext*) override;
  antlrcpp::Any visitSearchKeyDraft(imap::IMAPParser::SearchKeyDraftContext*) override;
  antlrcpp::Any visitSearchKeyHeader(imap::IMAPParser::SearchKeyHeaderContext*) override;
  antlrcpp::Any visitSearchKeyLarger(imap::IMAPParser::SearchKeyLargerContext*) override;
  antlrcpp::Any visitSearchKeyNot(imap::IMAPParser::SearchKeyNotContext*) override;
  antlrcpp::Any visitSearchKeyOr(imap::IMAPParser::SearchKeyOrContext*) override;
  antlrcpp::Any visitSearchKeySentBefore(imap::IMAPParser::SearchKeySentBeforeContext*) override;
  antlrcpp::Any visitSearchKeySentOn(imap::IMAPParser::SearchKeySentOnContext*) override;
  antlrcpp::Any visitSearchKeySentSince(imap::IMAPParser::SearchKeySentSinceContext*) override;
  antlrcpp::Any visitSearchKeySmaller(imap::IMAPParser::SearchKeySmallerContext*) override;
  antlrcpp::Any visitSearchKeyUID(imap::IMAPParser::SearchKeyUIDContext*) override;
  antlrcpp::Any visitSearchKeyUndraft(imap::IMAPParser::SearchKeyUndraftContext*) override;
  antlrcpp::Any visitSearchKeySeqSet(imap::IMAPParser::SearchKeySeqSetContext*) override;
  antlrcpp::Any visitSearchKeyList(imap::IMAPParser::SearchKeyListContext*) override;

  antlrcpp::Any visitDate(imap::IMAPParser::DateContext*) override;

  antlrcpp::Any visitSeqSet(imap::IMAPParser::SeqSetContext*) override;
  antlrcpp::Any visitSeqItemNum(imap::IMAPParser::SeqItemNumContext*) override;
  antlrcpp::Any visitSeqItemRng(imap::IMAPParser::SeqItemRngContext*) override;
  antlrcpp::Any visitSeqRange(imap::IMAPParser::SeqRangeContext*) override;

  antlrcpp::Any visitStoreAction(imap::IMAPParser::StoreActionContext*) override;

  antlrcpp::Any visitStoreFlagList(imap::IMAPParser::StoreFlagListContext*) override;
  antlrcpp::Any visitStoreSpacedFlags(imap::IMAPParser::StoreSpacedFlagsContext*) override;

  antlrcpp::Any visitFetchTargetAll(imap::IMAPParser::FetchTargetAllContext*) override;
  antlrcpp::Any visitFetchTargetFast(imap::IMAPParser::FetchTargetFastContext*) override;
  antlrcpp::Any visitFetchTargetFull(imap::IMAPParser::FetchTargetFullContext*) override;
  antlrcpp::Any visitFetchTargetAtt(imap::IMAPParser::FetchTargetAttContext*) override;

  antlrcpp::Any visitFetchAttEnvelope(imap::IMAPParser::FetchAttEnvelopeContext*) override;
  antlrcpp::Any visitFetchAttFlags(imap::IMAPParser::FetchAttFlagsContext*) override;
  antlrcpp::Any visitFetchAttInternalDate(imap::IMAPParser::FetchAttInternalDateContext*) override;
  antlrcpp::Any visitFetchAttRFC822(imap::IMAPParser::FetchAttRFC822Context*) override;
  antlrcpp::Any visitFetchAttRFC822Header(imap::IMAPParser::FetchAttRFC822HeaderContext*) override;
  antlrcpp::Any visitFetchAttRFC822Size(imap::IMAPParser::FetchAttRFC822SizeContext*) override;
  antlrcpp::Any visitFetchAttRFC822Text(imap::IMAPParser::FetchAttRFC822TextContext*) override;
  antlrcpp::Any visitFetchAttBody(imap::IMAPParser::FetchAttBodyContext*) override;
  antlrcpp::Any visitFetchAttBodyStructure(imap::IMAPParser::FetchAttBodyStructureContext*) override;
  antlrcpp::Any visitFetchAttUID(imap::IMAPParser::FetchAttUIDContext*) override;
  antlrcpp::Any visitFetchAttBodySection(imap::IMAPParser::FetchAttBodySectionContext*) override;

  antlrcpp::Any visitSectionKeyword(imap::IMAPParser::SectionKeywordContext*) override;
  antlrcpp::Any visitSectionWithPart(imap::IMAPParser::SectionWithPartContext*) override;
  antlrcpp::Any visitSectionKwdHeader(imap::IMAPParser::SectionKwdHeaderContext*) override;
  antlrcpp::Any visitSectionKwdHeaderFields(imap::IMAPParser::SectionKwdHeaderFieldsContext*) override;
  antlrcpp::Any visitSectionKwdText(imap::IMAPParser::SectionKwdTextContext*) override;
  antlrcpp::Any visitSectionText(imap::IMAPParser::SectionTextContext*) override;
  antlrcpp::Any visitPartial(imap::IMAPParser::PartialContext*) override;

  antlrcpp::Any visitNumber(imap::IMAPParser::NumberContext*) override;

 private:
  std::map<std::string, std::string> mLiterals;
};

}  // namespace parser

#endif
