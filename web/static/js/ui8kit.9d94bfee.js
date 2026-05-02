(()=>{function M(q){if(typeof document>"u"){q();return}if(document.readyState==="loading"){document.addEventListener("DOMContentLoaded",q,{once:!0});return}q()}function p(q,J=document){return J.querySelectorAll(`[data-${q}]`)}var t='a[href], button, input, select, textarea, [tabindex]:not([tabindex="-1"])';function v(q){return Array.from(q.querySelectorAll(t)).filter((z)=>!z.hasAttribute("disabled")&&z.getAttribute("aria-hidden")!=="true")}function w(){return typeof document<"u"&&typeof window<"u"}var q8="__UI8KIT_ARIA_AUTO_INIT__";function z8(){return globalThis[q8]!==!1}var _="ui8kit",Y={patterns:new Map,scopes:new Map},N=null,J8={ready:M,byAttr:p};function Q8(q){return q!=null?q:w()?document:{}}function W8(){let q=globalThis[_];if(q)return q;let J={core:J8,ready:M,byAttr:p,register:X,init:Z8,initPattern:D};if(w())globalThis[_]=J;return J}function x(){if(!N)N=W8();return N}function X(q){if(Y.patterns.has(q.name))return;if(Y.patterns.set(q.name,q),x()[q.name]=q,z8()&&w())M(()=>{D(q.name)})}function D(q,J){var z;let Q=Y.patterns.get(q);if(!Q)return;if(!w()&&!J)return;let W=Q8(J),Z=(z=Y.scopes.get(q))!=null?z:new WeakMap;if(Y.scopes.set(q,Z),Z.has(W))return;let $=Q.init(W);Z.set(W,typeof $==="function"?$:$8)}function Z8(q){for(let J of Y.patterns.values())D(J.name,q)}function $8(){}var P="data-state",c="open",j8="closed",G='[data-ui8kit="dialog"], [data-ui8kit="sheet"], [data-ui8kit="alertdialog"]',C='[data-ui8kit="dialog"][data-state="open"], [data-ui8kit="sheet"][data-state="open"], [data-ui8kit="alertdialog"][data-state="open"]',B=new WeakMap;function l(q){var J,z;if(!q)return null;if(typeof q==="string")return document.getElementById(q);let Q=q;if(typeof Q.matches==="function"&&Q.matches(G))return Q;let W=(J=Q.getAttribute)==null?void 0:J.call(Q,"data-ui8kit-dialog-target");if(W)return document.getElementById(W);return(z=Q.closest)==null?void 0:z.call(Q,G)}function L(q,J){if(!q)return;let z=q.querySelector("[data-ui8kit-dialog-overlay]");if(J){q.setAttribute(P,c),q.removeAttribute("hidden"),z==null||z.removeAttribute("hidden");let Q=v(q)[0];Q==null||Q.focus(),q.dataset.trapped="1"}else{q.setAttribute(P,j8),q.setAttribute("hidden","hidden"),z==null||z.setAttribute("hidden","hidden");let Q=B.get(q);if(Q&&document.contains(Q))Q.focus();B.delete(q),delete q.dataset.trapped}}function a(q){let J=l(q);if(!J)return;let z=q&&q instanceof HTMLElement&&q!==J?q:document.activeElement;if(z)B.set(J,z);L(J,!0)}function k(q){L(l(q),!1)}function E(q){if(q.key==="Escape"){let $=document.querySelector(C);if($)k($),q.preventDefault();return}if(q.key!=="Tab")return;let J=document.querySelector(C);if(!J||!J.contains(q.target))return;let z=v(J);if(z.length===0){q.preventDefault();return}let Q=z[0],W=z[z.length-1];if(!Q||!W){q.preventDefault();return}let Z=document.activeElement;if(q.shiftKey){if(Z===Q||Z===J)W.focus(),q.preventDefault()}else if(Z===W||Z===J)Q.focus(),q.preventDefault()}function O(q){let J=q.target;if(!J)return;let z=J.closest("[data-ui8kit-dialog-open]");if(z){a(z),q.preventDefault();return}if(J.closest("[data-ui8kit-dialog-close]")){k(J.closest(G)),q.preventDefault();return}if(J.closest("[data-ui8kit-dialog-overlay]"))k(J.closest(G))}var s={name:"dialog",init(q=document){let z=q.querySelectorAll(G);for(let Q of z)L(Q,Q.getAttribute(P)===c);return document.addEventListener("click",O),document.addEventListener("keydown",E),()=>{document.removeEventListener("click",O),document.removeEventListener("keydown",E)}},open:a,close:k};function I(q,J){let z=q.querySelector("[data-ui8kit-accordion-trigger]"),Q=q.querySelector("[data-ui8kit-accordion-content]");if(!z||!Q)return;if(q.setAttribute("data-state",J?"open":"closed"),z.setAttribute("aria-expanded",J?"true":"false"),J)Q.removeAttribute("hidden");else Q.setAttribute("hidden","hidden")}function H8(q){return q.getAttribute("data-accordion-type")==="multiple"}function X8(q,J){let z=q.querySelectorAll("[data-accordion-item]");for(let Q of z)if(Q!==J)I(Q,!1)}function U8(q){let J=q.querySelectorAll("[data-accordion-item]");for(let z of J){let Q=z.querySelector("[data-ui8kit-accordion-trigger]"),W=z.querySelector("[data-ui8kit-accordion-content]");if(!Q||!W||z.dataset.ui8kitBound)continue;z.dataset.ui8kitBound="1",Q.type="button",I(z,z.getAttribute("data-state")==="open"),Q.addEventListener("click",(Z)=>{Z.preventDefault();let $=z,V=z.getAttribute("data-state")!=="open";if(!H8(q))X8(q,$);I($,V)}),Q.addEventListener("keydown",(Z)=>{if(Z.key==="Enter"||Z.key===" ")Z.preventDefault(),Q.click()})}}var r={name:"accordion",init(q=document){let J=q.querySelectorAll('[data-ui8kit="accordion"]');for(let z of J)U8(z);return()=>{}}};function f(q,J,z){let Q=q.querySelectorAll("[data-tabs-trigger]"),W=q.querySelectorAll("[data-tabs-panel]");for(let Z of Q){let $=Z.getAttribute("data-tabs-value")===J;if(Z.setAttribute("aria-selected",$?"true":"false"),Z.tabIndex=$?0:-1,$&&z)Z.focus()}for(let Z of W)Z.hidden=Z.getAttribute("data-tabs-value")!==J}function Y8(q){var J;let z=q.getAttribute("data-tabs-value");if(z)return z;let Q=q.querySelector('[data-tabs-trigger][aria-selected="true"]');if(Q==null?void 0:Q.getAttribute("data-tabs-value"))return Q.getAttribute("data-tabs-value");return((J=q.querySelector("[data-tabs-trigger]"))==null?void 0:J.getAttribute("data-tabs-value"))||""}function G8(q,J){let z=Array.from(q.querySelectorAll("[data-tabs-trigger]"));if(!z.length)return;let Q=J.target.closest("[data-tabs-trigger]");if(!Q)return;let W=z.indexOf(Q);if(W<0)return;if(J.key==="ArrowRight"||J.key==="ArrowDown"){let Z=z[(W+1)%z.length];J.preventDefault(),f(q,Z.getAttribute("data-tabs-value")||"",!0)}else if(J.key==="ArrowLeft"||J.key==="ArrowUp"){let Z=z[(W-1+z.length)%z.length];J.preventDefault(),f(q,Z.getAttribute("data-tabs-value")||"",!0)}}var i={name:"tabs",init(q=document){let J=q.querySelectorAll('[data-ui8kit="tabs"]');for(let z of J){if(z.dataset.ui8kitBound)continue;z.dataset.ui8kitBound="1";let Q=Y8(z);if(Q)f(z,Q,!1);z.addEventListener("click",(W)=>{let Z=W.target.closest("[data-tabs-trigger]");if(!Z)return;let $=Z.getAttribute("data-tabs-value");if(!$)return;W.preventDefault(),f(z,$,!1)}),z.addEventListener("keydown",(W)=>{G8(z,W)})}return()=>{}}},F='[data-ui8kit="combobox"]';function U(q,J){let z=q.querySelector('[role="listbox"], ul'),Q=q.querySelector("[data-combobox-toggle]"),W=q.querySelector("input");if(!z||!W)return;if(q.dataset.state=J?"open":"closed",Q==null||Q.setAttribute("aria-expanded",String(J)),W.setAttribute("aria-expanded",String(J)),J)z.removeAttribute("hidden");else z.setAttribute("hidden","hidden")}function V8(q){return Array.from(q.querySelectorAll("[data-combobox-option]")).filter((z)=>{return z.style.display!=="none"&&z.getAttribute("aria-disabled")!=="true"})}function k8(q,J){let z=q.querySelectorAll("[data-combobox-option]");for(let Q of z)Q.classList.remove("ui-combobox-option-active"),Q.setAttribute("aria-selected","false");if(J)J.classList.add("ui-combobox-option-active"),J.setAttribute("aria-selected","true")}function R(q){let J=q.querySelector("input");if(!J)return;let z=q.querySelectorAll("[data-combobox-option]"),Q=J.value.trim().toLowerCase();for(let W of z){let Z=(W.textContent||"").toLowerCase(),$=Q.length===0||Z.includes(Q);W.style.display=$?"":"none"}}function b(q,J){if(!J)return;let z=q.querySelector("input");if(!z)return;let Q=J.getAttribute("data-combobox-value")||J.textContent||"";z.value=Q,U(q,!1)}function f8(q){let J=q.querySelector("input"),z=q.querySelector('[role="listbox"], ul'),Q=q.querySelector("[data-combobox-toggle]");if(!J||!z)return()=>{};let W=()=>U(q,!0),Z=()=>{U(q,!0),R(q)},$=(j)=>{let H=V8(q);if(j.key==="ArrowDown"||j.key==="ArrowUp"){if(H.length===0)return;let y=q.querySelector(".ui-combobox-option-active"),S=y?H.indexOf(y):-1,A=j.key==="ArrowDown"?(S+1)%H.length:(S-1+H.length)%H.length;k8(q,H[A]);let K=H[A];if(K&&typeof K.scrollIntoView==="function")try{K.scrollIntoView({block:"nearest"})}catch(w8){}j.preventDefault();return}if(j.key==="Enter"){b(q,q.querySelector(".ui-combobox-option-active")),j.preventDefault();return}if(j.key==="Escape")U(q,!1),j.preventDefault()},V=()=>{let j=q.dataset.state==="open";U(q,!j)},T=(j)=>j.preventDefault(),h=[];for(let j of q.querySelectorAll("[data-combobox-option]")){let H=()=>b(q,j);j.addEventListener("mousedown",T),j.addEventListener("click",H),h.push({el:j,handler:H})}return J.addEventListener("focus",W),J.addEventListener("input",Z),J.addEventListener("keydown",$),Q==null||Q.addEventListener("click",V),R(q),()=>{J.removeEventListener("focus",W),J.removeEventListener("input",Z),J.removeEventListener("keydown",$),Q==null||Q.removeEventListener("click",V);for(let{el:j,handler:H}of h)j.removeEventListener("mousedown",T),j.removeEventListener("click",H)}}function m(q){var J;let z=q.target;if(!z||((J=z.closest)==null?void 0:J.call(z,F)))return;for(let Q of document.querySelectorAll(F))U(Q,!1)}var n={name:"combobox",init(q=document){let z=q.querySelectorAll(F),Q=[];for(let W of z){if(W.dataset.ui8kitBound)continue;W.dataset.ui8kitBound="1",Q.push(f8(W))}return document.addEventListener("click",m),()=>{document.removeEventListener("click",m);for(let W of Q)W()}}};function d(q){let J=q.querySelector('[role="tooltip"]');if(!J)return;J.removeAttribute("hidden"),q.setAttribute("data-state","open"),J.setAttribute("aria-hidden","false")}function u(q){let J=q.querySelector('[role="tooltip"]');if(!J)return;J.setAttribute("hidden","hidden"),q.setAttribute("data-state","closed"),J.setAttribute("aria-hidden","true")}var o={name:"tooltip",init(q=document){let J=q.querySelectorAll('[data-ui8kit="tooltip"]');for(let z of J){if(z.dataset.ui8kitBound)continue;z.addEventListener("mouseenter",()=>d(z)),z.addEventListener("focusin",()=>d(z)),z.addEventListener("mouseleave",()=>u(z)),z.addEventListener("focusout",()=>u(z)),z.dataset.ui8kitBound="1"}return()=>{}}};function g(q,J){let z=q.querySelector("[data-disclosure-trigger]"),Q=q.querySelector("[data-disclosure-content]");if(!z||!Q)return;if(z.setAttribute("aria-expanded",String(J)),J)Q.removeAttribute("hidden");else Q.setAttribute("hidden","hidden")}var e={name:"disclosure",init(q=document){let J=q.querySelectorAll('[data-ui8kit="disclosure"]');for(let z of J){let Q=z.querySelector("[data-disclosure-trigger]"),W=z.querySelector("[data-disclosure-content]");if(!Q||!W||z.dataset.ui8kitBound)continue;let Z=Q.getAttribute("aria-expanded")==="true";g(z,Z),Q.addEventListener("click",()=>{let $=Q.getAttribute("aria-expanded")!=="true";g(z,$)}),Q.addEventListener("keydown",($)=>{if($.key==="Enter"||$.key===" ")$.preventDefault(),Q.click()}),z.dataset.ui8kitBound="1"}return()=>{}}};X(r);X(n);X(s);X(e);X(i);X(o);if(typeof window<"u"&&window.__UI8KIT_ARIA_AUTO_INIT__!==!1)x().init();})();
;
(function () {
  var namespace = window.ui8kit || {};
  window.ui8kit = namespace;
  if (namespace.languageSwitch) {
    return;
  }

  function ready(fn) {
    if (document.readyState === "loading") {
      document.addEventListener("DOMContentLoaded", fn);
      return;
    }
    fn();
  }

  function parseResponse(html) {
    var parser = new DOMParser();
    return parser.parseFromString(html, "text/html");
  }

  function replaceMainContent(button, html) {
    var targetSelector = button.getAttribute("data-spa-target") || "main";
    var parsed = parseResponse(html);

    var currentTarget = document.querySelector(targetSelector);
    var nextTarget = parsed.querySelector(targetSelector);
    if (currentTarget && nextTarget) {
      currentTarget.innerHTML = nextTarget.innerHTML;
    }

    var parsedTitle = parsed.querySelector("title");
    if (parsedTitle && parsedTitle.textContent) {
      document.title = parsedTitle.textContent;
    }

    var nextButton = button.id ? parsed.getElementById(button.id) : null;
    if (nextButton) {
      if (nextButton.getAttribute("href")) {
        button.setAttribute("href", nextButton.getAttribute("href"));
      }
      if (nextButton.dataset.currentLocale) {
        button.dataset.currentLocale = nextButton.dataset.currentLocale;
      }
      if (nextButton.dataset.nextLocale) {
        button.dataset.nextLocale = nextButton.dataset.nextLocale;
      }
      button.textContent = nextButton.textContent;
    }

    var locale = parsed.documentElement && parsed.documentElement.getAttribute("lang");
    if (locale) {
      document.documentElement.setAttribute("lang", locale);
    }
  }

  function bindLanguageSwitch(button) {
    button.addEventListener("click", function (event) {
      event.preventDefault();
      var href = button.getAttribute("href");
      if (!href) {
        return;
      }

      fetch(href, {
        credentials: "same-origin",
        headers: {
          "X-Locale-Switch": "1",
        },
      })
        .then(function (response) {
          if (!response.ok) {
            throw new Error("locale switch request failed");
          }
          return response.text();
        })
        .then(function (html) {
          replaceMainContent(button, html);
          history.pushState({}, "", href);
        })
        .catch(function () {
          window.location.href = href;
        });
    });
  }

  ready(function () {
    var toggles = document.querySelectorAll("[data-ui8kit-spa-lang]");
    for (var i = 0; i < toggles.length; i += 1) {
      bindLanguageSwitch(toggles[i]);
    }
  });

  namespace.languageSwitch = { init: function () {} };
})();
