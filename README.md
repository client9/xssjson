XssJSON escapes all strings inside an existing JSON document

Before:

```json
{
"key": "<script>alert(\"1\")</script>"
}
```

After

```json
{
"key": "&lt;script&gt;alert(&quot;1&quot;)&lt;&#x2fscript&lt;"
}
```

(not so sure about escaping forward slash but thats how it is for now.)

## Whats Escaped

From [OWASP XSS Prevention Cheat Sheet](https://www.owasp.org/index.php/XSS_(Cross_Site_Scripting)_Prevention_Cheat_Sheet):

```
 & --> &amp;
 < --> &lt;
 > --> &gt;
 " --> &quot;
 ' --> &#x27;     &apos; not recommended because its not in the HTML spec (See: section 24.4.1) &apos; is in the XML and XHTML specs.
 / --> &#x2F;     forward slash is included as it helps end an HTML entity
```

Although not actually clear what the attack is with forward slash.  Oh well.


## Pros and Cons

Pros:
* works on the json stream, does not need to know about the objects or uses reflection.
* any json encoder can be used.  This just filters the result.


Cons:
* performance would be best if the json encoder just did this natively.
