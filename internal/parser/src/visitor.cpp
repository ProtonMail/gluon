#include "visitor.h"
#include <algorithm>
#include <iostream>

namespace parser {

Visitor::Visitor(const std::map<std::string, std::string>& literals) : mLiterals{literals} {}

antlrcpp::Any Visitor::visitTag(imap::IMAPParser::TagContext* ctx) {
  return ctx->getText();
}

antlrcpp::Any Visitor::visitStringQtd(imap::IMAPParser::StringQtdContext* ctx) {
  auto res = std::string{};

  for (const auto& c : ctx->quoted()->quotedChar())
    res.append(visit(c).as<std::string>());

  return res;
}

antlrcpp::Any Visitor::visitStringLit(imap::IMAPParser::StringLitContext* ctx) {
  return visit(ctx->literal());
}

antlrcpp::Any Visitor::visitRawQChar(imap::IMAPParser::RawQCharContext* ctx) {
  return ctx->getText();
}

antlrcpp::Any Visitor::visitEscQChar(imap::IMAPParser::EscQCharContext* ctx) {
  return ctx->quotedSpecial()->getText();
}

antlrcpp::Any Visitor::visitLiteral(imap::IMAPParser::LiteralContext* ctx) {
  return mLiterals.at(ctx->uuid()->getText());
}

antlrcpp::Any Visitor::visitAstringRaw(imap::IMAPParser::AstringRawContext* ctx) {
  return ctx->getText();
}

antlrcpp::Any Visitor::visitAstringStr(imap::IMAPParser::AstringStrContext* ctx) {
  return visit(ctx->string());
}

antlrcpp::Any Visitor::visitCommand(imap::IMAPParser::CommandContext* ctx) {
  proto::Command command;

  if (ctx->commandAny()) {
    auto cmd = ctx->commandAny();

    if (cmd->capability()) {
      command.set_allocated_capability(visit(cmd->capability()).as<proto::Capability*>());
    } else if (cmd->noop()) {
      command.set_allocated_noop(visit(cmd->noop()).as<proto::Noop*>());
    } else if (cmd->logout()) {
      command.set_allocated_logout(visit(cmd->logout()).as<proto::Logout*>());
    }
  } else if (ctx->commandNonAuth()) {
    auto cmd = ctx->commandNonAuth();

    if (cmd->auth()) {
      command.set_allocated_auth(visit(cmd->auth()).as<proto::Auth*>());
    } else if (cmd->startTLS()) {
      command.set_allocated_starttls(visit(cmd->startTLS()).as<proto::StartTLS*>());
    } else if (cmd->login()) {
      command.set_allocated_login(visit(cmd->login()).as<proto::Login*>());
    }
  } else if (ctx->commandAuth()) {
    auto cmd = ctx->commandAuth();

    if (cmd->select()) {
      command.set_allocated_select(visit(cmd->select()).as<proto::Select*>());
    } else if (cmd->examine()) {
      command.set_allocated_examine(visit(cmd->examine()).as<proto::Examine*>());
    } else if (cmd->create()) {
      command.set_allocated_create(visit(cmd->create()).as<proto::Create*>());
    } else if (cmd->del()) {
      command.set_allocated_del(visit(cmd->del()).as<proto::Del*>());
    } else if (cmd->rename()) {
      command.set_allocated_rename(visit(cmd->rename()).as<proto::Rename*>());
    } else if (cmd->sub()) {
      command.set_allocated_sub(visit(cmd->sub()).as<proto::Sub*>());
    } else if (cmd->unsub()) {
      command.set_allocated_unsub(visit(cmd->unsub()).as<proto::Unsub*>());
    } else if (cmd->list()) {
      command.set_allocated_list(visit(cmd->list()).as<proto::List*>());
    } else if (cmd->lsub()) {
      command.set_allocated_lsub(visit(cmd->lsub()).as<proto::Lsub*>());
    } else if (cmd->status()) {
      command.set_allocated_status(visit(cmd->status()).as<proto::Status*>());
    } else if (cmd->append()) {
      command.set_allocated_append(visit(cmd->append()).as<proto::Append*>());
    } else if (cmd->idle()) {
      command.set_allocated_idle(visit(cmd->idle()).as<proto::Idle*>());
    }
  } else if (ctx->commandSelected()) {
    auto cmd = ctx->commandSelected();

    if (cmd->check()) {
      command.set_allocated_check(visit(cmd->check()).as<proto::Check*>());
    } else if (cmd->close()) {
      command.set_allocated_close(visit(cmd->close()).as<proto::Close*>());
    } else if (cmd->expunge()) {
      command.set_allocated_expunge(visit(cmd->expunge()).as<proto::Expunge*>());
    } else if (cmd->uidExpunge()) {
      command.set_allocated_uidexpunge(visit(cmd->uidExpunge()).as<proto::UIDExpunge*>());
    } else if (cmd->unselect()) {
      command.set_allocated_unselect(visit(cmd->unselect()).as<proto::Unselect*>());
    } else if (cmd->search()) {
      command.set_allocated_search(visit(cmd->search()).as<proto::Search*>());
    } else if (cmd->fetch()) {
      command.set_allocated_fetch(visit(cmd->fetch()).as<proto::Fetch*>());
    } else if (cmd->store()) {
      command.set_allocated_store(visit(cmd->store()).as<proto::Store*>());
    } else if (cmd->copy()) {
      command.set_allocated_copy(visit(cmd->copy()).as<proto::Copy*>());
    } else if (cmd->move()) {
      command.set_allocated_move(visit(cmd->move()).as<proto::Move*>());
    } else if (cmd->uid()) {
      command.set_allocated_uid(visit(cmd->uid()).as<proto::UID*>());
    }
  } else if (ctx->done()) {
    command.set_allocated_done(visit(ctx->done()).as<proto::Done*>());
  }

  return command;
}

antlrcpp::Any Visitor::visitCapability(imap::IMAPParser::CapabilityContext*) {
  return new proto::Capability;
}

antlrcpp::Any Visitor::visitNoop(imap::IMAPParser::NoopContext*) {
  return new proto::Noop;
}

antlrcpp::Any Visitor::visitLogout(imap::IMAPParser::LogoutContext*) {
  return new proto::Logout;
}

antlrcpp::Any Visitor::visitStartTLS(imap::IMAPParser::StartTLSContext*) {
  return new proto::StartTLS;
}

antlrcpp::Any Visitor::visitAuth(imap::IMAPParser::AuthContext* ctx) {
  auto auth = new proto::Auth;

  auth->set_type(ctx->authType()->getText());

  for (const auto& base64 : ctx->base64()) {
    auth->add_data(base64->getText());
  }

  return auth;
}

antlrcpp::Any Visitor::visitLogin(imap::IMAPParser::LoginContext* ctx) {
  auto login = new proto::Login;

  login->set_username(visit(ctx->userID()->astring()).as<std::string>());
  login->set_password(visit(ctx->password()->astring()).as<std::string>());

  return login;
}

antlrcpp::Any Visitor::visitMboxInbox(imap::IMAPParser::MboxInboxContext* ctx) {
  return ctx->getText();
}

antlrcpp::Any Visitor::visitMboxOther(imap::IMAPParser::MboxOtherContext* ctx) {
  return visit(ctx->astring());
}

antlrcpp::Any Visitor::visitListMboxRaw(imap::IMAPParser::ListMboxRawContext* ctx) {
  return ctx->getText();
}

antlrcpp::Any Visitor::visitListMboxStr(imap::IMAPParser::ListMboxStrContext* ctx) {
  return visit(ctx->string());
}

antlrcpp::Any Visitor::visitSelect(imap::IMAPParser::SelectContext* ctx) {
  auto select = new proto::Select;

  select->set_mailbox(visit(ctx->mailbox()).as<std::string>());

  return select;
}

antlrcpp::Any Visitor::visitExamine(imap::IMAPParser::ExamineContext* ctx) {
  auto examine = new proto::Examine;

  examine->set_mailbox(visit(ctx->mailbox()).as<std::string>());

  return examine;
}

antlrcpp::Any Visitor::visitCreate(imap::IMAPParser::CreateContext* ctx) {
  auto create = new proto::Create;

  create->set_mailbox(visit(ctx->mailbox()).as<std::string>());

  return create;
}

antlrcpp::Any Visitor::visitDel(imap::IMAPParser::DelContext* ctx) {
  auto del = new proto::Del;

  del->set_mailbox(visit(ctx->mailbox()).as<std::string>());

  return del;
}

antlrcpp::Any Visitor::visitRename(imap::IMAPParser::RenameContext* ctx) {
  auto rename = new proto::Rename;

  rename->set_mailbox(visit(ctx->mailbox(0)).as<std::string>());
  rename->set_newname(visit(ctx->mailbox(1)).as<std::string>());

  return rename;
}

antlrcpp::Any Visitor::visitSub(imap::IMAPParser::SubContext* ctx) {
  auto sub = new proto::Sub;

  sub->set_mailbox(visit(ctx->mailbox()).as<std::string>());

  return sub;
}

antlrcpp::Any Visitor::visitUnsub(imap::IMAPParser::UnsubContext* ctx) {
  auto unsub = new proto::Unsub;

  unsub->set_mailbox(visit(ctx->mailbox()).as<std::string>());

  return unsub;
}

antlrcpp::Any Visitor::visitList(imap::IMAPParser::ListContext* ctx) {
  auto list = new proto::List;

  list->set_reference(visit(ctx->mailbox()).as<std::string>());
  list->set_mailbox(visit(ctx->listMailbox()).as<std::string>());

  return list;
}

antlrcpp::Any Visitor::visitLsub(imap::IMAPParser::LsubContext* ctx) {
  auto lsub = new proto::Lsub;

  lsub->set_reference(visit(ctx->mailbox()).as<std::string>());
  lsub->set_mailbox(visit(ctx->listMailbox()).as<std::string>());

  return lsub;
}

antlrcpp::Any Visitor::visitStatus(imap::IMAPParser::StatusContext* ctx) {
  auto status = new proto::Status;

  status->set_mailbox(visit(ctx->mailbox()).as<std::string>());

  for (const auto& att : ctx->statusAtt())
    status->add_attributes(att->getText());

  return status;
}

antlrcpp::Any Visitor::visitAppend(imap::IMAPParser::AppendContext* ctx) {
  auto append = new proto::Append;

  append->set_mailbox(visit(ctx->mailbox()).as<std::string>());

  append->set_message(visit(ctx->literal()).as<std::string>());

  if (ctx->dateTime())
    append->set_allocated_datetime(visit(ctx->dateTime()).as<proto::DateTime*>());

  if (ctx->flagList()) {
    auto flags = visit(ctx->flagList()).as<std::vector<std::string>>();

    for (const auto& flag : flags)
      append->add_flags(flag);
  }

  return append;
}

antlrcpp::Any Visitor::visitCheck(imap::IMAPParser::CheckContext* ctx) {
  return new proto::Check;
}

antlrcpp::Any Visitor::visitClose(imap::IMAPParser::CloseContext* ctx) {
  return new proto::Close;
}

antlrcpp::Any Visitor::visitExpunge(imap::IMAPParser::ExpungeContext* ctx) {
  return new proto::Expunge;
}

antlrcpp::Any Visitor::visitUidExpunge(imap::IMAPParser::UidExpungeContext* ctx) {
  auto expunge = new proto::UIDExpunge;

  expunge->set_allocated_sequenceset(visit(ctx->seqSet()).as<proto::SequenceSet*>());

  return expunge;
}

antlrcpp::Any Visitor::visitUnselect(imap::IMAPParser::UnselectContext* ctx) {
  return new proto::Unselect;
}

antlrcpp::Any Visitor::visitSearch(imap::IMAPParser::SearchContext* ctx) {
  auto search = new proto::Search;

  if (ctx->astring())
    search->set_charset(visit(ctx->astring()).as<std::string>());

  for (const auto& key : ctx->searchKey())
    search->add_keys()->CopyFrom(*visit(key).as<proto::SearchKey*>());

  return search;
}

antlrcpp::Any Visitor::visitFetch(imap::IMAPParser::FetchContext* ctx) {
  auto fetch = new proto::Fetch;

  fetch->set_allocated_sequenceset(visit(ctx->seqSet()).as<proto::SequenceSet*>());

  auto atts = visit(ctx->fetchTarget()).as<std::vector<proto::FetchAttribute*>>();

  for (const auto& att : atts)
    fetch->add_attributes()->CopyFrom(*att);

  return fetch;
}

antlrcpp::Any Visitor::visitStore(imap::IMAPParser::StoreContext* ctx) {
  auto store = new proto::Store;

  store->set_allocated_sequenceset(visit(ctx->seqSet()).as<proto::SequenceSet*>());

  store->set_allocated_action(visit(ctx->storeAction()).as<proto::StoreAction*>());

  auto flags = visit(ctx->storeFlags()).as<std::vector<std::string>>();

  for (const auto& flag : flags)
    store->add_flags(flag);

  return store;
}

antlrcpp::Any Visitor::visitCopy(imap::IMAPParser::CopyContext* ctx) {
  auto copy = new proto::Copy;

  copy->set_allocated_sequenceset(visit(ctx->seqSet()).as<proto::SequenceSet*>());

  copy->set_mailbox(visit(ctx->mailbox()).as<std::string>());

  return copy;
}

antlrcpp::Any Visitor::visitMove(imap::IMAPParser::MoveContext* ctx) {
  auto move = new proto::Move;

  move->set_allocated_sequenceset(visit(ctx->seqSet()).as<proto::SequenceSet*>());

  move->set_mailbox(visit(ctx->mailbox()).as<std::string>());

  return move;
}

antlrcpp::Any Visitor::visitUid(imap::IMAPParser::UidContext* ctx) {
  auto uid = new proto::UID;

  if (ctx->copy()) {
    uid->set_allocated_copy(visit(ctx->copy()).as<proto::Copy*>());
  }

  if (ctx->fetch()) {
    uid->set_allocated_fetch(visit(ctx->fetch()).as<proto::Fetch*>());
  }

  if (ctx->search()) {
    uid->set_allocated_search(visit(ctx->search()).as<proto::Search*>());
  }

  if (ctx->store()) {
    uid->set_allocated_store(visit(ctx->store()).as<proto::Store*>());
  }

  if (ctx->move()) {
    uid->set_allocated_move(visit(ctx->move()).as<proto::Move*>());
  }

  return uid;
}

antlrcpp::Any Visitor::visitIdle(imap::IMAPParser::IdleContext* ctx) {
  return new proto::Idle;
}

antlrcpp::Any Visitor::visitDone(imap::IMAPParser::DoneContext* ctx) {
  return new proto::Done;
}

antlrcpp::Any Visitor::visitFlagList(imap::IMAPParser::FlagListContext* ctx) {
  auto flags = std::vector<std::string>{};

  for (const auto& flag : ctx->flag())
    flags.push_back(flag->getText());

  return flags;
}

antlrcpp::Any Visitor::visitDateTime(imap::IMAPParser::DateTimeContext* ctx) {
  auto dateTime = new proto::DateTime;

  dateTime->mutable_date()->set_day(std::atoi(ctx->dateDayFixed()->getText().c_str()));

  auto month = ctx->dateMonth()->getText();

  std::transform(month.begin(), month.end(), month.begin(), [](auto c) { return std::tolower(c); });

  if (month == "jan") {
    dateTime->mutable_date()->set_month(1);
  } else if (month == "feb") {
    dateTime->mutable_date()->set_month(2);
  } else if (month == "mar") {
    dateTime->mutable_date()->set_month(3);
  } else if (month == "apr") {
    dateTime->mutable_date()->set_month(4);
  } else if (month == "may") {
    dateTime->mutable_date()->set_month(5);
  } else if (month == "jun") {
    dateTime->mutable_date()->set_month(6);
  } else if (month == "jul") {
    dateTime->mutable_date()->set_month(7);
  } else if (month == "aug") {
    dateTime->mutable_date()->set_month(8);
  } else if (month == "sep") {
    dateTime->mutable_date()->set_month(9);
  } else if (month == "oct") {
    dateTime->mutable_date()->set_month(10);
  } else if (month == "nov") {
    dateTime->mutable_date()->set_month(11);
  } else if (month == "dec") {
    dateTime->mutable_date()->set_month(12);
  }

  dateTime->mutable_date()->set_year(std::atoi(ctx->dateYear()->getText().c_str()));

  dateTime->set_allocated_time(visit(ctx->time()).as<proto::Time*>());

  dateTime->set_allocated_zone(visit(ctx->zone()).as<proto::Zone*>());

  return dateTime;
}

antlrcpp::Any Visitor::visitTime(imap::IMAPParser::TimeContext* ctx) {
  auto time = new proto::Time;

  auto hour = ctx->digit(0)->getText() + ctx->digit(1)->getText();
  auto minute = ctx->digit(2)->getText() + ctx->digit(3)->getText();
  auto second = ctx->digit(4)->getText() + ctx->digit(5)->getText();

  time->set_hour(std::atoi(hour.c_str()));
  time->set_minute(std::atoi(minute.c_str()));
  time->set_second(std::atoi(second.c_str()));

  return time;
}

antlrcpp::Any Visitor::visitZone(imap::IMAPParser::ZoneContext* ctx) {
  auto zone = new proto::Zone;

  auto hour = ctx->digit(0)->getText() + ctx->digit(1)->getText();
  auto minute = ctx->digit(2)->getText() + ctx->digit(3)->getText();

  zone->set_hour(std::atoi(hour.c_str()));
  zone->set_minute(std::atoi(minute.c_str()));
  zone->set_sign(visit(ctx->sign()).as<bool>());

  return zone;
}

antlrcpp::Any Visitor::visitSignPlus(imap::IMAPParser::SignPlusContext*) {
  return true;
}

antlrcpp::Any Visitor::visitSignMinus(imap::IMAPParser::SignMinusContext*) {
  return false;
}

antlrcpp::Any Visitor::visitSearchKeyAll(imap::IMAPParser::SearchKeyAllContext*) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWAll);

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyAnswered(imap::IMAPParser::SearchKeyAnsweredContext*) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWAnswered);

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyBcc(imap::IMAPParser::SearchKeyBccContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWBcc);

  key->set_text(visit(ctx->astring()).as<std::string>());

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyBefore(imap::IMAPParser::SearchKeyBeforeContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWBefore);

  key->set_date(visit(ctx->date()).as<std::string>());

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyBody(imap::IMAPParser::SearchKeyBodyContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWBody);

  key->set_text(visit(ctx->astring()).as<std::string>());

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyCc(imap::IMAPParser::SearchKeyCcContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWCc);

  key->set_text(visit(ctx->astring()).as<std::string>());

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyDeleted(imap::IMAPParser::SearchKeyDeletedContext*) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWDeleted);

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyFlagged(imap::IMAPParser::SearchKeyFlaggedContext*) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWFlagged);

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyFrom(imap::IMAPParser::SearchKeyFromContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWFrom);

  key->set_text(visit(ctx->astring()).as<std::string>());

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyKeyword(imap::IMAPParser::SearchKeyKeywordContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWKeyword);

  key->set_flag(ctx->flagKeyword()->getText());

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyNew(imap::IMAPParser::SearchKeyNewContext*) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWNew);

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyOld(imap::IMAPParser::SearchKeyOldContext*) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWOld);

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyOn(imap::IMAPParser::SearchKeyOnContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWOn);

  key->set_date(visit(ctx->date()).as<std::string>());

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyRecent(imap::IMAPParser::SearchKeyRecentContext*) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWRecent);

  return key;
}

antlrcpp::Any Visitor::visitSearchKeySeen(imap::IMAPParser::SearchKeySeenContext*) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWSeen);

  return key;
}

antlrcpp::Any Visitor::visitSearchKeySince(imap::IMAPParser::SearchKeySinceContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWSince);

  key->set_date(visit(ctx->date()).as<std::string>());

  return key;
}

antlrcpp::Any Visitor::visitSearchKeySubject(imap::IMAPParser::SearchKeySubjectContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWSubject);

  key->set_text(visit(ctx->astring()).as<std::string>());

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyText(imap::IMAPParser::SearchKeyTextContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWText);

  key->set_text(visit(ctx->astring()).as<std::string>());

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyTo(imap::IMAPParser::SearchKeyToContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWTo);

  key->set_text(visit(ctx->astring()).as<std::string>());

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyUnanswered(imap::IMAPParser::SearchKeyUnansweredContext*) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWUnanswered);

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyUndeleted(imap::IMAPParser::SearchKeyUndeletedContext*) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWUndeleted);

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyUnflagged(imap::IMAPParser::SearchKeyUnflaggedContext*) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWUnflagged);

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyUnkeyword(imap::IMAPParser::SearchKeyUnkeywordContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWUnkeyword);

  key->set_flag(ctx->flagKeyword()->getText());

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyUnseen(imap::IMAPParser::SearchKeyUnseenContext*) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWUnseen);

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyDraft(imap::IMAPParser::SearchKeyDraftContext*) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWDraft);

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyHeader(imap::IMAPParser::SearchKeyHeaderContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWHeader);

  key->set_text(visit(ctx->astring()).as<std::string>());

  key->set_field(visit(ctx->headerFieldName()->astring()).as<std::string>());

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyLarger(imap::IMAPParser::SearchKeyLargerContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWLarger);

  key->set_size(visit(ctx->number()).as<int>());

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyNot(imap::IMAPParser::SearchKeyNotContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWNot);

  key->set_allocated_leftop(visit(ctx->searchKey()).as<proto::SearchKey*>());

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyOr(imap::IMAPParser::SearchKeyOrContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWOr);

  key->set_allocated_leftop(visit(ctx->searchKey()[0]).as<proto::SearchKey*>());

  key->set_allocated_rightop(visit(ctx->searchKey()[1]).as<proto::SearchKey*>());

  return key;
}

antlrcpp::Any Visitor::visitSearchKeySentBefore(imap::IMAPParser::SearchKeySentBeforeContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWSentBefore);

  key->set_date(visit(ctx->date()).as<std::string>());

  return key;
}

antlrcpp::Any Visitor::visitSearchKeySentOn(imap::IMAPParser::SearchKeySentOnContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWSentOn);

  key->set_date(visit(ctx->date()).as<std::string>());

  return key;
}

antlrcpp::Any Visitor::visitSearchKeySentSince(imap::IMAPParser::SearchKeySentSinceContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWSentSince);

  key->set_date(visit(ctx->date()).as<std::string>());

  return key;
}

antlrcpp::Any Visitor::visitSearchKeySmaller(imap::IMAPParser::SearchKeySmallerContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWSmaller);

  key->set_size(visit(ctx->number()).as<int>());

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyUID(imap::IMAPParser::SearchKeyUIDContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWUID);

  key->set_allocated_sequenceset(visit(ctx->seqSet()).as<proto::SequenceSet*>());

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyUndraft(imap::IMAPParser::SearchKeyUndraftContext*) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWUndraft);

  return key;
}

antlrcpp::Any Visitor::visitSearchKeySeqSet(imap::IMAPParser::SearchKeySeqSetContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWSeqSet);

  key->set_allocated_sequenceset(visit(ctx->seqSet()).as<proto::SequenceSet*>());

  return key;
}

antlrcpp::Any Visitor::visitSearchKeyList(imap::IMAPParser::SearchKeyListContext* ctx) {
  auto key = new proto::SearchKey;

  key->set_keyword(proto::SearchKeyword::SearchKWList);

  for (const auto& child : ctx->searchKey())
    key->add_children()->CopyFrom(*visit(child).as<proto::SearchKey*>());

  return key;
}

antlrcpp::Any Visitor::visitDate(imap::IMAPParser::DateContext* ctx) {
  return ctx->dateText()->getText();
}

antlrcpp::Any Visitor::visitSeqSet(imap::IMAPParser::SeqSetContext* ctx) {
  auto sequenceSet = new proto::SequenceSet;

  for (const auto& item : ctx->seqItem())
    sequenceSet->add_items()->CopyFrom(*visit(item).as<proto::SequenceItem*>());

  return sequenceSet;
}

antlrcpp::Any Visitor::visitSeqItemNum(imap::IMAPParser::SeqItemNumContext* ctx) {
  auto seqItem = new proto::SequenceItem;

  seqItem->set_number(ctx->seqNumber()->getText());

  return seqItem;
}

antlrcpp::Any Visitor::visitSeqItemRng(imap::IMAPParser::SeqItemRngContext* ctx) {
  auto seqItem = new proto::SequenceItem;

  seqItem->set_allocated_range(visit(ctx->seqRange()).as<proto::SequenceRange*>());

  return seqItem;
}

antlrcpp::Any Visitor::visitSeqRange(imap::IMAPParser::SeqRangeContext* ctx) {
  auto seqRange = new proto::SequenceRange;

  seqRange->set_begin(ctx->seqNumber(0)->getText());
  seqRange->set_end(ctx->seqNumber(1)->getText());

  return seqRange;
}

antlrcpp::Any Visitor::visitStoreAction(imap::IMAPParser::StoreActionContext* ctx) {
  auto storeAction = new proto::StoreAction;

  if (ctx->sign()) {
    if (visit(ctx->sign()).as<bool>())
      storeAction->set_operation(proto::Operation::Add);
    else
      storeAction->set_operation(proto::Operation::Remove);
  } else {
    storeAction->set_operation(proto::Operation::Replace);
  }

  if (ctx->silent())
    storeAction->set_silent(true);

  return storeAction;
}

antlrcpp::Any Visitor::visitStoreFlagList(imap::IMAPParser::StoreFlagListContext* ctx) {
  return visit(ctx->flagList());
}

antlrcpp::Any Visitor::visitStoreSpacedFlags(imap::IMAPParser::StoreSpacedFlagsContext* ctx) {
  auto flags = std::vector<std::string>{};

  for (const auto& flag : ctx->flag())
    flags.push_back(flag->getText());

  return flags;
}

// ALL macro is equivalent to FLAGS INTERNALDATE RFC822.SIZE ENVELOPE.
antlrcpp::Any Visitor::visitFetchTargetAll(imap::IMAPParser::FetchTargetAllContext* ctx) {
  auto flags = new proto::FetchAttribute;
  flags->set_keyword(proto::FetchKeyword::FetchKWFlags);

  auto date = new proto::FetchAttribute;
  date->set_keyword(proto::FetchKeyword::FetchKWInternalDate);

  auto size = new proto::FetchAttribute;
  size->set_keyword(proto::FetchKeyword::FetchKWRFC822Size);

  auto env = new proto::FetchAttribute;
  env->set_keyword(proto::FetchKeyword::FetchKWEnvelope);

  return std::vector<proto::FetchAttribute*>{flags, date, size, env};
}

// FAST macro is equivalent to FLAGS INTERNALDATE RFC822.SIZE.
antlrcpp::Any Visitor::visitFetchTargetFast(imap::IMAPParser::FetchTargetFastContext* ctx) {
  auto flags = new proto::FetchAttribute;
  flags->set_keyword(proto::FetchKeyword::FetchKWFlags);

  auto date = new proto::FetchAttribute;
  date->set_keyword(proto::FetchKeyword::FetchKWInternalDate);

  auto size = new proto::FetchAttribute;
  size->set_keyword(proto::FetchKeyword::FetchKWRFC822Size);

  return std::vector<proto::FetchAttribute*>{flags, date, size};
}

// FULL macro is equivalent to FLAGS INTERNALDATE RFC822.SIZE ENVELOPE BODY.
antlrcpp::Any Visitor::visitFetchTargetFull(imap::IMAPParser::FetchTargetFullContext* ctx) {
  auto flags = new proto::FetchAttribute;
  flags->set_keyword(proto::FetchKeyword::FetchKWFlags);

  auto date = new proto::FetchAttribute;
  date->set_keyword(proto::FetchKeyword::FetchKWInternalDate);

  auto size = new proto::FetchAttribute;
  size->set_keyword(proto::FetchKeyword::FetchKWRFC822Size);

  auto env = new proto::FetchAttribute;
  env->set_keyword(proto::FetchKeyword::FetchKWEnvelope);

  auto body = new proto::FetchAttribute;
  body->set_keyword(proto::FetchKeyword::FetchKWBody);

  return std::vector<proto::FetchAttribute*>{flags, date, size, env, body};
}

antlrcpp::Any Visitor::visitFetchTargetAtt(imap::IMAPParser::FetchTargetAttContext* ctx) {
  auto atts = std::vector<proto::FetchAttribute*>{};

  for (const auto& att : ctx->fetchAtt())
    atts.push_back(visit(att).as<proto::FetchAttribute*>());

  return atts;
}

antlrcpp::Any Visitor::visitFetchAttEnvelope(imap::IMAPParser::FetchAttEnvelopeContext*) {
  auto att = new proto::FetchAttribute;

  att->set_keyword(proto::FetchKeyword::FetchKWEnvelope);

  return att;
}

antlrcpp::Any Visitor::visitFetchAttFlags(imap::IMAPParser::FetchAttFlagsContext*) {
  auto att = new proto::FetchAttribute;

  att->set_keyword(proto::FetchKeyword::FetchKWFlags);

  return att;
}

antlrcpp::Any Visitor::visitFetchAttInternalDate(imap::IMAPParser::FetchAttInternalDateContext*) {
  auto att = new proto::FetchAttribute;

  att->set_keyword(proto::FetchKeyword::FetchKWInternalDate);

  return att;
}

antlrcpp::Any Visitor::visitFetchAttRFC822(imap::IMAPParser::FetchAttRFC822Context* ctx) {
  auto att = new proto::FetchAttribute;

  att->set_keyword(proto::FetchKeyword::FetchKWRFC822);

  return att;
}

antlrcpp::Any Visitor::visitFetchAttRFC822Header(imap::IMAPParser::FetchAttRFC822HeaderContext* ctx) {
  auto att = new proto::FetchAttribute;

  att->set_keyword(proto::FetchKeyword::FetchKWRFC822Header);

  return att;
}

antlrcpp::Any Visitor::visitFetchAttRFC822Size(imap::IMAPParser::FetchAttRFC822SizeContext* ctx) {
  auto att = new proto::FetchAttribute;

  att->set_keyword(proto::FetchKeyword::FetchKWRFC822Size);

  return att;
}

antlrcpp::Any Visitor::visitFetchAttRFC822Text(imap::IMAPParser::FetchAttRFC822TextContext* ctx) {
  auto att = new proto::FetchAttribute;

  att->set_keyword(proto::FetchKeyword::FetchKWRFC822Text);

  return att;
}

antlrcpp::Any Visitor::visitFetchAttBody(imap::IMAPParser::FetchAttBodyContext*) {
  auto att = new proto::FetchAttribute;

  att->set_keyword(proto::FetchKeyword::FetchKWBody);

  return att;
}

antlrcpp::Any Visitor::visitFetchAttBodyStructure(imap::IMAPParser::FetchAttBodyStructureContext*) {
  auto att = new proto::FetchAttribute;

  att->set_keyword(proto::FetchKeyword::FetchKWBodyStructure);

  return att;
}

antlrcpp::Any Visitor::visitFetchAttUID(imap::IMAPParser::FetchAttUIDContext*) {
  auto att = new proto::FetchAttribute;

  att->set_keyword(proto::FetchKeyword::FetchKWUID);

  return att;
}

antlrcpp::Any Visitor::visitFetchAttBodySection(imap::IMAPParser::FetchAttBodySectionContext* ctx) {
  auto att = new proto::FetchAttribute;

  // We always set the body object, even if it's empty,
  // because an empty body means to fetch the full message.
  att->set_allocated_body(new proto::FetchBody);

  if (ctx->peek())
    att->mutable_body()->set_peek(true);

  if (ctx->section())
    att->mutable_body()->set_allocated_section(visit(ctx->section()).as<proto::BodySection*>());

  if (ctx->partial())
    att->mutable_body()->set_allocated_partial(visit(ctx->partial()).as<proto::BodyPartial*>());

  return att;
}

antlrcpp::Any Visitor::visitSectionKeyword(imap::IMAPParser::SectionKeywordContext* ctx) {
  return visit(ctx->sectionMsgText());
}

antlrcpp::Any Visitor::visitSectionKwdHeader(imap::IMAPParser::SectionKwdHeaderContext*) {
  auto section = new proto::BodySection;

  section->set_keyword(proto::SectionKeyword::Header);

  return section;
}

antlrcpp::Any Visitor::visitSectionKwdHeaderFields(imap::IMAPParser::SectionKwdHeaderFieldsContext* ctx) {
  auto section = new proto::BodySection;

  if (!ctx->headerFieldsNot())
    section->set_keyword(proto::SectionKeyword::HeaderFields);
  else
    section->set_keyword(proto::SectionKeyword::HeaderFieldsNot);

  for (const auto& field : ctx->headerList()->headerFieldName())
    section->add_fields(visit(field->astring()).as<std::string>());

  return section;
}

antlrcpp::Any Visitor::visitSectionKwdText(imap::IMAPParser::SectionKwdTextContext*) {
  auto section = new proto::BodySection;

  section->set_keyword(proto::SectionKeyword::Text);

  return section;
}

antlrcpp::Any Visitor::visitSectionText(imap::IMAPParser::SectionTextContext* ctx) {
  if (ctx->sectionMsgText())
    return visit(ctx->sectionMsgText());

  auto section = new proto::BodySection;

  section->set_keyword(proto::SectionKeyword::MIME);

  return section;
}

antlrcpp::Any Visitor::visitSectionWithPart(imap::IMAPParser::SectionWithPartContext* ctx) {
  auto section = new proto::BodySection;

  if (ctx->sectionText())
    section->CopyFrom(*visit(ctx->sectionText()).as<proto::BodySection*>());

  for (const auto& number : ctx->sectionPart()->number())
    section->add_parts(visit(number).as<int>());

  return section;
}

antlrcpp::Any Visitor::visitPartial(imap::IMAPParser::PartialContext* ctx) {
  auto partial = new proto::BodyPartial;

  partial->set_begin(visit(ctx->number(0)).as<int>());
  partial->set_count(visit(ctx->number(1)).as<int>());

  return partial;
}

antlrcpp::Any Visitor::visitNumber(imap::IMAPParser::NumberContext* ctx) {
  return std::atoi(ctx->getText().c_str());
}

}  // namespace parser
