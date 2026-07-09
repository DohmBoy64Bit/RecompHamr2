// Package skills loads embedded and custom RecompHamr skill documents.
//
// The package owns embedded skill inventory, custom skill discovery,
// case-insensitive resolution, active-list rendering, audit classification, and
// the network-free parts of the skill-new workflow. Callers that fetch remote
// skill text must pass the fetched body into NewDraft, cache it for review, and
// use ScaffoldCustomSkill only after explicit approval.
package skills
