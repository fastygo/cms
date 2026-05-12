function safeMenuLocationAttr(value) {
  var raw = String(value || "").trim();
  if (!/^[a-zA-Z0-9_-]+$/.test(raw)) {
    return "";
  }
  return raw;
}

function syncLocaleMenuRegions(parsed) {
  var sections = document.querySelectorAll("[data-gocms-menu-location]");
  for (var i = 0; i < sections.length; i++) {
    var el = sections[i];
    var loc = safeMenuLocationAttr(el.getAttribute("data-gocms-menu-location"));
    if (!loc) {
      continue;
    }
    var selector = '[data-gocms-menu-location="' + loc + '"]';
    var sameInDoc = document.querySelectorAll(selector);
    var sameInParsed = parsed.querySelectorAll(selector);
    var idx = Array.prototype.indexOf.call(sameInDoc, el);
    if (idx >= 0 && idx < sameInParsed.length) {
      el.innerHTML = sameInParsed[idx].innerHTML;
    }
  }
}
