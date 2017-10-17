package main

import (
	"time"
)

// Digest allows simple protection of hex formatted digest strings, prefixed
// by their algorithm. Strings of type Digest have some guarantee of being in
// the correct format and it provides quick access to the components of a
// digest string.
//
// The following is an example of the contents of Digest types:
//
// 	sha256:7173b809ca12ec5dee4506cd86be934c4596dd234ee82c0662eac04a8c2c71dc
//
// This allows to abstract the digest behind this type and work only in those
// terms.
// Copied from github.com/opencontainers/go-digest/digest.go
type Digest string

// RegistryNotification defines the fields of a json event envelope message that can hold
// one or more events. Definitions are copied from https://github.com/docker/distribution/blob/master/notifications/event.go#L42
type RegistryNotification struct {
	Events []RegistryEvent `json:"events"`
}

// RegistryEvent provides the fields required to describe a registry event.
type RegistryEvent struct {
	// ID provides a unique identifier for the event.
	ID string `json:"id,omitempty"`

	// Timestamp is the time at which the event occurred.
	Timestamp time.Time `json:"timestamp,omitempty"`

	// Action indicates what action encompasses the provided event.
	Action string `json:"action,omitempty"`

	// Target uniquely describes the target of the event.
	Target struct {
		// TODO(stevvooe): Use http.DetectContentType for layers, maybe.

		Descriptor

		// Length in bytes of content. Same as Size field in Descriptor.
		// Provided for backwards compatibility.
		Length int64 `json:"length,omitempty"`

		// Repository identifies the named repository.
		Repository string `json:"repository,omitempty"`

		// FromRepository identifies the named repository which a blob was mounted
		// from if appropriate.
		FromRepository string `json:"fromRepository,omitempty"`

		// URL provides a direct link to the content.
		URL string `json:"url,omitempty"`

		// Tag provides the tag
		Tag string `json:"tag,omitempty"`
	} `json:"target,omitempty"`

	// Request covers the request that generated the event.
	Request RequestRecord `json:"request,omitempty"`

	// Actor specifies the agent that initiated the event. For most
	// situations, this could be from the authorization context of the request.
	Actor ActorRecord `json:"actor,omitempty"`

	// Source identifies the registry node that generated the event. Put
	// differently, while the actor "initiates" the event, the source
	// "generates" it.
	Source SourceRecord `json:"source,omitempty"`
}

// Descriptor describes targeted content. Used in conjunction with a blob
// store, a descriptor can be used to fetch, store and target any kind of
// blob. The struct also describes the wire protocol format. Fields should
// only be added but never changed.
type Descriptor struct {
	// MediaType describe the type of the content. All text based formats are
	// encoded as utf-8.
	MediaType string `json:"mediaType,omitempty"`

	// Size in bytes of content.
	Size int64 `json:"size,omitempty"`

	// Digest uniquely identifies the content. A byte stream can be verified
	// against against this digest.
	Digest Digest `json:"digest,omitempty"`

	// URLs contains the source URLs of this content.
	URLs []string `json:"urls,omitempty"`
}

// RequestRecord covers the request that generated the event.
type RequestRecord struct {
	// ID uniquely identifies the request that initiated the event.
	ID string `json:"id"`

	// Addr contains the ip or hostname and possibly port of the client
	// connection that initiated the event. This is the RemoteAddr from
	// the standard http request.
	Addr string `json:"addr,omitempty"`

	// Host is the externally accessible host name of the registry instance,
	// as specified by the http host header on incoming requests.
	Host string `json:"host,omitempty"`

	// Method has the request method that generated the event.
	Method string `json:"method"`

	// UserAgent contains the user agent header of the request.
	UserAgent string `json:"useragent"`
}

// ActorRecord specifies the agent that initiated the event. For most
// situations, this could be from the authorization context of the request.
// Data in this record can refer to both the initiating client and the
// generating request.
type ActorRecord struct {
	// Name corresponds to the subject or username associated with the
	// request context that generated the event.
	Name string `json:"name,omitempty"`
}

// SourceRecord identifies the registry node that generated the event. Put
// differently, while the actor "initiates" the event, the source "generates"
// it.
type SourceRecord struct {
	// Addr contains the ip or hostname and the port of the registry node
	// that generated the event. Generally, this will be resolved by
	// os.Hostname() along with the running port.
	Addr string `json:"addr,omitempty"`

	// InstanceID identifies a running instance of an application. Changes
	// after each restart.
	InstanceID string `json:"instanceID,omitempty"`
}
