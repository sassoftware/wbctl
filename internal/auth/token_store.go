// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/sassoftware/sas-ggdk/pkg/errors"
	"github.com/sassoftware/sas-ggdk/pkg/jsonutils"
	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/spf13/afero"
)

// timeNow is hook for testing to override the time.Now() func.
var timeNow = time.Now

// tokenCache represents a collection of all cached tokens, keyed by the
// endpoint they are associated with.
type tokenCache map[string]cachedToken

// cachedToken represents the on disk format for cached tokens.
type cachedToken struct {
	AccessToken  string    `json:"access_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
}

func getValidAccessTokenFromCachedToken(ct cachedToken) (string, error) {
	if ct.AccessToken == "" {
		return "", errors.New(i18n.Format(i18n.NoAccessTokenInCache))
	}
	// Give some grace for time-of-check vs time-of-use for this token.
	if time.Now().After(ct.ExpiresAt.Add(-1 * time.Minute)) {
		return "", errors.New(i18n.Format(i18n.CachedAccessTokenIsExpired))
	}
	return ct.AccessToken, nil
}

func loadTokenCache(fsys afero.Fs, path string) (tokenCache, error) {
	fs := afero.Afero{Fs: fsys}
	res, err := result.Split(jsonutils.LoadWith[tokenCache](fs.ReadFile, path))
	if err != nil {
		if os.IsNotExist(err) {
			// If the file doesn't exist, return an empty cache instead of an error.
			return tokenCache{}, nil
		}
		return nil, errors.Wrap(err, i18n.Format(i18n.FailedToLoadTokenCache))
	}
	return res, nil
}

// getCachedTokenFor loads the cached token from ~/.wbctl/token.json.
func getCachedTokenFor(fsys afero.Fs, endpoint string) (cachedToken, error) {
	path, err := getTokenCachePath()
	if err != nil {
		return cachedToken{}, err
	}
	cache, err := loadTokenCache(fsys, path)
	if err != nil {
		return cachedToken{}, err
	}
	return getTokenFromCache(endpoint, cache), nil
}

// saveTokenToCache converts the given token to the cached representation and saves
// that cached representation to ~/.wbctl/token.json.
func saveTokenToCache(fsys afero.Fs, endpoint string, responseToken tokenResponse) error {
	path, err := getTokenCachePath()
	if err != nil {
		return errors.Wrap(err, i18n.Format(i18n.FailedToSaveTokenCache))
	}
	cache, err := loadTokenCache(fsys, path)
	if err != nil {
		return errors.Wrap(err, i18n.Format(i18n.FailedToSaveTokenCache))
	}
	ct := cachedToken{
		AccessToken:  responseToken.AccessToken,
		RefreshToken: responseToken.RefreshToken,
		TokenType:    responseToken.TokenType,
		ExpiresAt:    timeNow().UTC().Add(time.Duration(responseToken.ExpiresIn) * time.Second),
	}
	endpoint = normalizeEndpoint(endpoint)
	cache[endpoint] = ct
	dir := filepath.Dir(path)
	err = fsys.MkdirAll(dir, 0o700)
	if err != nil {
		return errors.Wrap(err, i18n.Format(i18n.FailedToSaveTokenCache))
	}
	err = saveCache(fsys, cache, path)
	if err != nil {
		return errors.Wrap(err, i18n.Format(i18n.FailedToSaveTokenCache))
	}
	return nil
}

// getTokenCachePath returns the full path to the token cache file.
func getTokenCachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, constants.DefaultTokenFileName), nil
}

// saveCache returns a function that saves the given cache to the
// specified path with the given permissions. It uses a temp file + rename to
// ensure that the file is not left in a corrupted state if the save is
// interrupted.
func saveCache(
	fsys afero.Fs,
	cache tokenCache,
	p string,
) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	tmp, err := afero.TempFile(
		fsys,
		path.Join(homeDir, constants.DefaultConfigDirName),
		constants.DefaultTokenTmpFileName,
	)
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	err = tmp.Close()
	if err != nil {
		return err
	}
	err = jsonutils.SaveWith(afero.Afero{Fs: fsys}.WriteFile, cache, tmpName, 0o600)
	if err != nil {
		_ = fsys.Remove(tmpName)
		return err
	}
	err = fsys.Rename(tmpName, p)
	if err != nil {
		_ = fsys.Remove(tmpName)
		return err
	}
	return nil
}

func getTokenFromCache(
	endpoint string,
	cache tokenCache,
) cachedToken {
	endpoint = normalizeEndpoint(endpoint)
	found, present := cache[endpoint]
	if !present {
		return cachedToken{}
	}
	return found
}

func normalizeEndpoint(endpoint string) string {
	parsedEndpoint, err := url.Parse(endpoint)
	if err == nil {
		parsedEndpoint.Path = ""
		parsedEndpoint.RawQuery = ""
		parsedEndpoint.Fragment = ""
		endpoint = parsedEndpoint.String()
	}
	return endpoint
}
