package artex

import (
 "encoding/json"
 "strings"
 "testing"
)

func TestEvidenceCanonicalization(t *testing.T) {
 a,redacted,err:=CanonicalEvidence(json.RawMessage(`{"z":9007199254740993,"a":{"password":"sensitive","API-Key":"other","nested":[{"token":"hidden","note":"keep"}]}}`))
 if err!=nil || !redacted { t.Fatal("canonicalization/redaction failed") }
 for _,secret:=range []string{"sensitive","other","hidden"} { if strings.Contains(string(a),secret) { t.Fatal("credential retained") } }
 b,_,err:=CanonicalEvidence(json.RawMessage(`{ "a":{"nested":[{"note":"keep","token":"hidden"}],"API-Key":"other","password":"sensitive"}, "z":9007199254740993 }`))
 if err!=nil || string(a)!=string(b) || !strings.Contains(string(a),"9007199254740993") { t.Fatal("unstable canonical bytes or numeric precision lost") }
 for _,raw:=range []string{"null","[]","true",`{"a":1} {"b":2}`,strings.Repeat("x",MaxEvidenceBytes+1)} { if _,_,err=CanonicalEvidence(json.RawMessage(raw));err==nil { t.Fatal("accepted invalid evidence") } }
}

func TestEvidenceRecordsUseSourceIDs(t *testing.T) {
 cases:=[]struct{body,kind,parent,id string}{
  {`{"assets":[{"id":9007199254740993,"type":"ip"}]}`,"asset","","9007199254740993"},
  {`{"items":[{"id":"123","evidence":"original"}]}`,"finding","","123"},
  {`{"items":[{"seq":121,"summary":"observed"}],"cursor":121}`,"activity","","121"},
  {`{"retests":[{"id":12,"finding_id":123,"verdict":"inconclusive"}]}`,"retest","123","12"},
 }
 for _,tt:=range cases {
  t.Run(tt.kind,func(t *testing.T){
   records,err:=EvidenceRecords(json.RawMessage(tt.body),tt.kind,"list",tt.parent)
   if err!=nil || len(records)!=1 || records[0].ID!=tt.id || records[0].ParentID!=tt.parent { t.Fatal("lost source identity",err) }
  })
 }
 if _,err:=EvidenceRecords(json.RawMessage(`{"retests":[{"id":12,"finding_id":999}]}`),"retest","list","123");err==nil { t.Fatal("accepted mismatched parent") }
 for _,body:=range []string{`{}`,`{"items":[{"id":0}]}`,`{"items":"wrong"}`} { if _,err:=EvidenceRecords(json.RawMessage(body),"finding","list","");err==nil { t.Fatal("accepted invalid collection") } }
 if EvidenceTaskID(json.RawMessage(`{"task_id":123}`))!="123" || EvidenceTaskID(json.RawMessage(`{"task_id":"456"}`))!="456" { t.Fatal("task ID parsing failed") }
}
