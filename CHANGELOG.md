<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Client Breaking" for breaking CLI commands and REST routes used by end-users.
"API Breaking" for breaking exported APIs used by developers building on SDK.
"State Machine Breaking" for any changes that result in a different AppState
given same genesisState and txList.
Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## v0.0.20

- Revert: Fix nonce tracker first fetch. Mark the first fetch as done so that it is incremented in IncrementAndGet()

## v0.0.19

- Fix nonce tracker first fetch. Mark the first fetch as done so that it is incremented in IncrementAndGet()

- Add max duration to async request processor, change it from relying on retry config.

## v0.0.18

- Async request processor abstraction

## v0.0.17

- Add Cosmos signer.

## v0.0.16

- Allow case sensitive header keys in HTTP requests.

## v0.0.15

- Add GetCurrentNonce() and ForceUpdateNonce() to NonceTrackerMock.

## v0.0.14

- Add GetCurrentNonce() and ForceUpdateNonce() to nonce tracker.

## v0.0.13

- Nonce tracker mock.

## v0.0.12 

- Add scaling factor precompute for fast math operations.

## v0.0.10

- Add GetLastFailureTime and GetLastSuccessTime to circuit breaker.

## v0.0.9

- Add OnError callback to circuit breaker.

## v0.0.8

- Add circuit breaker pattern implementation.

## v0.0.7

- Add GetVenueAssets API to SwapVenueI.

## v0.0.6

- Add mock swap venue.

## v0.0.5

- Return interface from BinanceSwapVenue constructor.

## v0.0.4

- Omit zero balances from GetBalances API.

## v0.0.3

- GetBalances API now supports optional denom filtering.

## v0.0.2

- Add Binance swap venue.
- Add tx package with nonce tracker.

## v0.0.1

-  Initial release with httputil package.
