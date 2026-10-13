//go:build artexintegration

package artex

import (
 "context"
 "crypto/sha256"
 "database/sql"
 "encoding/hex"
 "encoding/json"
 "errors"
 "net/url"
 "os"
 "strings"
 "sync"
 "testing"
 "time"

 "github.com/google/uuid"
 _ "github.com/lib/pq"
 "pentagi/migrations"
)

func evidenceLabDB(t *testing.T) *sql.DB {
 t.Helper()
 dsn:=os.Getenv("ARTEX_LAB_PG_DSN")
 endpoint,err:=url.Parse(dsn)
 if err!=nil || dsn=="" || endpoint.Hostname()!="postgres" || endpoint.Path!="/artex" { t.Fatal("isolated ARTEX lab database is required") }
 root,err:=sql.Open("postgres",dsn); if err!=nil { t.Fatal("open lab failed") }
 t.Cleanup(func(){root.Close()})
 schema:="pentagi_evidence_"+strings.ReplaceAll(uuid.NewString(),"-","")
 if _,err=root.Exec("CREATE SCHEMA "+schema);err!=nil{t.Fatal("create lab schema failed")}
 t.Cleanup(func(){ if _,err:=root.Exec("DROP SCHEMA "+schema+" CASCADE");err!=nil{t.Error("lab schema cleanup failed")} })
 query:=endpoint.Query();query.Set("search_path",schema);endpoint.RawQuery=query.Encode()
 db,err:=sql.Open("postgres",endpoint.String());if err!=nil{t.Fatal("open schema failed")}
 t.Cleanup(func(){db.Close()});db.SetMaxOpenConns(8)
 for _,name:=range []string{"sql/20260925120000_engine_runs.sql","sql/20260925130000_engine_evidence_snapshots.sql"} {
  migration,err:=migrations.EmbedMigrations.ReadFile(name);if err!=nil{t.Fatal(err)}
  if _,err=db.Exec(strings.Split(string(migration),"-- +goose Down")[0]);err!=nil{t.Fatal("migration failed",err)}
 }
 return db
}

func TestSQLEvidenceHistoryLive(t *testing.T) {
 db:=evidenceLabDB(t)
 ctx,cancel:=context.WithTimeout(context.Background(),45*time.Second);defer cancel()
 store:=NewSQLRunStore(db)
 run,err:=store.CreatePending(ctx,7,RunRequest{Name:"evidence fixture",BindingID:"local-fixture",BindingRevision:"v1"})
 if err!=nil{t.Fatal(err)}
 if err=store.MarkCreated(ctx,run.ID,"123");err!=nil{t.Fatal(err)}
 record:=EvidenceRecord{ID:"9007199254740993",Kind:"finding",Representation:"detail",Content:json.RawMessage(`{"id":"9007199254740993","evidence":"original","api_key":"synthetic-secret"}`)}
 var wg sync.WaitGroup
 items:=make(chan *EvidenceSnapshot,8);failures:=make(chan error,8)
 for i:=0;i<8;i++{wg.Add(1);go func(){defer wg.Done();item,err:=store.SaveEvidence(ctx,7,run.ID,"123",record);items<-item;failures<-err}()}
 wg.Wait();close(items);close(failures)
 for err:=range failures{if err!=nil{t.Fatal("concurrent save failed",err)}}
 var original *EvidenceSnapshot
 for item:=range items{if original!=nil && (item.ID!=original.ID || !item.CapturedAt.Equal(original.CapturedAt) || item.Version!=1){t.Fatal("duplicate observation mutated immutable metadata")};original=item}
 t.Log("PASS concurrent_duplicate_keeps_original_id_version_and_time")
 if !original.Redacted || strings.Contains(string(original.Content),"synthetic-secret"){t.Fatal("credential archive leak")}
 fetched,err:=NewSQLRunStore(db).GetSnapshot(ctx,run.ID,original.ID)
 if err!=nil{t.Fatal(err)}
 digest:=sha256.Sum256(fetched.Content);if hex.EncodeToString(digest[:])!=fetched.ContentSHA256{t.Fatal("stored bytes differ from fingerprint")}
 t.Log("PASS persisted_content_hash_and_redaction")
 if _,err=store.SaveEvidence(ctx,8,run.ID,"123",record);!errors.Is(err,ErrRunNotFound){t.Fatal("foreign owner accepted")}
 if _,err=store.SaveEvidence(ctx,7,run.ID,"999",record);!errors.Is(err,ErrRunNotFound){t.Fatal("foreign task accepted")}
 t.Log("PASS owner_and_remote_task_match_required")
 changed:=record;changed.Content=json.RawMessage(`{"id":"9007199254740993","evidence":"changed"}`)
 second,err:=store.SaveEvidence(ctx,7,run.ID,"123",changed);if err!=nil || second.Version!=2{t.Fatal("version two failed",err)}
 reverted,err:=store.SaveEvidence(ctx,7,run.ID,"123",record);if err!=nil || reverted.Version!=3 || reverted.ID==original.ID{t.Fatal("reversion lost",err)}
 record.Representation="list"
 summary,err:=store.SaveEvidence(ctx,7,run.ID,"123",record);if err!=nil || summary.Version!=1{t.Fatal("list overwrote detail",err)}
 t.Log("PASS changed_and_reverted_content_preserves_history_separate_from_lists")
 retest:=EvidenceRecord{ID:"44",ParentID:record.ID,Kind:"retest",Representation:"list",Content:json.RawMessage(`{"id":44,"finding_id":"9007199254740993","verdict":"inconclusive"}`)}
 retestItem,err:=store.SaveEvidence(ctx,7,run.ID,"123",retest);if err!=nil || retestItem.ParentRecordID!=record.ID{t.Fatal("retest provenance failed",err)}
 originalAfter,_:=store.GetSnapshot(ctx,run.ID,original.ID);if string(originalAfter.Content)!=string(original.Content){t.Fatal("retest overwrote evidence")}
 t.Log("PASS_retest_links_finding_without_overwriting_source")
 page,err:=store.SnapshotPage(ctx,run.ID,"","",2);if err!=nil || !page.HasMore || len(page.Items)!=2{t.Fatal("pagination failed",err)}
 if len(page.Items[0].Content)!=0{t.Fatal("unbounded content in metadata list")}
 observed:=map[string]bool{};for _,item:=range page.Items{observed[item.ID]=true}
 record.ID="88";record.Content=json.RawMessage(`{"id":88,"evidence":"later arrival"}`)
 added,err:=store.SaveEvidence(ctx,7,run.ID,"123",record);if err!=nil{t.Fatal(err)}
 for page.HasMore{
  page,err=store.SnapshotPage(ctx,run.ID,"",page.NextCursor,2);if err!=nil{t.Fatal(err)}
  for _,item:=range page.Items{if observed[item.ID] || item.ID==added.ID{t.Fatal("cursor replay or new arrival shifted page")};observed[item.ID]=true}
 }
 if len(observed)!=5{t.Fatal("lost records across concurrent insertion",len(observed))}
 t.Log("PASS cursor_pagination_survives_new_observations")
 if _,err=store.GetSnapshot(ctx,uuid.NewString(),original.ID);!errors.Is(err,ErrSnapshotNotFound){t.Fatal("snapshot crossed run boundary")}
}
