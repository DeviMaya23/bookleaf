export type LinkPermalinkRule = {
  id: string;
  matches: (url: URL) => boolean;
};

const TWITTER_HOSTNAMES = new Set(["twitter.com", "x.com", "www.twitter.com", "www.x.com"]);
const TWITTER_STATUS_PATH = /\/status\/\d+/;

const twitterRule: LinkPermalinkRule = {
  id: "twitter-status-permalink",
  matches: (url) => TWITTER_HOSTNAMES.has(url.hostname) && TWITTER_STATUS_PATH.test(url.pathname),
};

const FACEBOOK_HOSTNAMES = new Set(["facebook.com", "www.facebook.com", "m.facebook.com"]);
const FACEBOOK_POST_PATH = /\/(posts|photo|photos|permalink\.php|story\.php)/;

const facebookRule: LinkPermalinkRule = {
  id: "facebook-post-permalink",
  matches: (url) => FACEBOOK_HOSTNAMES.has(url.hostname) && FACEBOOK_POST_PATH.test(url.pathname),
};

export const rules: LinkPermalinkRule[] = [twitterRule, facebookRule];

export function resolveLinkPermalink(linkUrl: string): boolean {
  let url: URL;
  try {
    url = new URL(linkUrl);
  } catch {
    return false;
  }
  return rules.some((rule) => rule.matches(url));
}
