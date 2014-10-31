package osmapi

import (
	"encoding/xml"
	"errors"
	"log"
)

/* ===
Changesets
*/

type ChangeSetSt struct {
	Id      string
	Request *MyRequestSt
	OsmCh   *OsmChangeSt
	//Type   string     // node, way rel
}

type ChangeSt struct {
	//XMLName xml.Name `xml:",omitempty"`
	Node     []*NodeSt   `xml:"node,omitempty"`
	Way      *WaySt      `xml:"way,omitempty"`
	Relation *RelationSt `xml:"relation,omitempty"`
}

type OsmChangeSt struct {
	XMLName    xml.Name   `xml:"osmChange"`
	Version    string     `xml:"version,attr"`
	Generator  string     `xml:"generator,attr"`
	Modify     *ChangeSt  `xml:"modify,omitempty"`
	Create     *ChangeSt  `xml:"create,omitempty"`
	Delete     *ChangeSt  `xml:"delete,omitempty"`
	Changeset  *TagListSt `xml:"changeset,omitempty"`
	Type       string     `xml:"-"`
	ChangeType string     `xml:"-"`
}

type OsmSt struct {
	XMLName   xml.Name   `xml:"osm"`
	Version   string     `xml:"version,attr"`
	Generator string     `xml:"generator,attr"`
	Changeset *TagListSt `xml:"changeset,omitempty"`
}

func (r *MyRequestSt) Changesets(t string) (*ChangeSetSt, error) {
	c := ChangeSetSt{}

	c.Id = ""
	c.Request = r
	if err := c.Create(); err != nil {
		return nil, err
	}

	if err := c.OsmChange(t); err != nil {
		return nil, err
	}

	return &c, nil
}

/*   */
func (ChSet *ChangeSetSt) OsmChange(t string) error {
	OsmCh := OsmChangeSt{}

	if t != "create" && t != "modify" && t != "delete" && t != "changeset" {
		return errors.New("OsmChange. You have to use create|modify|delete as OsmChange type. Now it is " + t)
	}

	OsmCh.Type = t
	OsmCh.Version = ProtocolVersion
	OsmCh.Generator = UserAgent
	ch := ChangeSt{[]*NodeSt{}, nil, nil}

	switch OsmCh.Type {
	case "modify":
		OsmCh.Modify = &ch
	case "create":
		OsmCh.Create = &ch
	case "delete":
		OsmCh.Delete = &ch
	}

	ChSet.OsmCh = &OsmCh

	return nil
}

/* ===
Changesets: Close: PUT /api/0.6/changeset/#id/close
*/
func (ChSet *ChangeSetSt) Close() error {
	/*  Changesets doesn't open. May by :) */
	if ChSet.Id == "" {
		return nil
	}

	/* Answer has to be empty */
	_, err := ChSet.Request.Put("/api/0.6/changeset/" + ChSet.Id + "/close")

	/* Clean memory. For any case */
	ChSet.OsmCh = nil
	ChSet = nil

	return err
}

/* ===
Changesets: Create: PUT /api/0.6/changeset/create
*/
func (ChSet *ChangeSetSt) Create() error {

	t := OsmSt{}
	t.Version = "0.6"
	t.Generator = UserAgent
	t.Changeset = &TagListSt{[]*TagSt{NewTag("comment", "changeset comment"), NewTag("created_by", UserAgent)}}
	body2, err2 := xml.MarshalIndent(t, "", "")
	if err2 != nil {
		return err2
	}

	ChSet.Id = ""
	id, err := ChSet.Request.Put("/api/0.6/changeset/create", string(body2))
	if err == nil {
		ChSet.Id = id
	}

	if id == "" {
		return errors.New("Bad answer. Data from " + ChSet.Request.Url + " does not contain changeset's id.")
	}

	return err
}

/* ===
Changesets: Diff upload: POST /api/0.6/changeset/#id/upload
*/
func (ChSet *ChangeSetSt) Upload() (string, error) {

	//(c *ChangeSetSt)
	if ChSet.Id == "" {
		errors.New("Cann't use uninitialize")
	}

	body, err_m := xml.MarshalIndent(ChSet.OsmCh, "", "")
	if err_m != nil {
		return "", err_m
	}

	data, err := ChSet.Request.PostXML("/api/0.6/changeset/"+ChSet.Id+"/upload", string(body))
	if err != nil {
		return "", err
	}

	log.Println("REEEEEEEED: /diffResult/" + ChSet.OsmCh.ChangeType + "/@new_id")

	old_id := xml_str(data, "/diffResult/"+ChSet.OsmCh.ChangeType+"/@old_id")
	new_id := xml_str(data, "/diffResult/"+ChSet.OsmCh.ChangeType+"/@new_id")

	err_line := "Bad result ChangeSetSt upload."

	if ChSet.OsmCh.Type == "modify" && old_id != new_id {
		return "", errors.New(err_line + "Old " + ChSet.OsmCh.ChangeType + " id equals new.")
	}

	if ChSet.OsmCh.Type == "delete" && "0" != new_id && "" != new_id {
		return "", errors.New(err_line + " Delete " + ChSet.OsmCh.ChangeType + ". Bad new id for " + ChSet.OsmCh.Type)
	}

	if ChSet.OsmCh.Type == "create" && ("" == new_id || "0" == new_id) {
		return "", errors.New(err_line + " Create. New " + ChSet.OsmCh.ChangeType + " id empty for " + ChSet.OsmCh.Type)
	}

	return new_id, nil
}
