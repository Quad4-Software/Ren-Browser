-∿
<

`c`!Hello!`! This is output from `*micron`*
Micron generates formatted text for your terminal
`a


-∿
<


Nomad Network supports a simple and functional markup language called `*micron`*. If you are familiar with `*markdown`* or `*HTML`*, you will feel right at home writing pages with micron.

With micron you can easily create structured documents and pages with formatting, colors, glyphs and icons, ideal for display in terminals.

>>Recommendations and Requirements

While micron can output formatted text to even the most basic terminal, there's a few capabilities your terminal `*must`* support to display micron output correctly, and some that, while not strictly necessary, make the experience a lot better.

Formatting such as `_underline`_, `!bold`! or `*italics`* will be displayed if your terminal supports it.

If you are having trouble getting micron output to display correctly, try using `*gnome-terminal`* or `*alacritty`*, which should work with all formatting options out of the box. Most other terminals will work fine as well, but you might have to change some settings to get certain formatting to display correctly.

>>>Encoding

All micron sources are intepreted as UTF-8, and micron assumes it can output UTF-8 characters to the terminal. If your terminal does not support UTF-8, output will be faulty.

>>>Colors

Shading and coloring text and backgrounds is integral to micron output, and while micron will attempt to gracefully degrade output even to 1-bit terminals, you will get the best output with terminals supporting at least 256 colors. True-color support is recommended.

>>>Terminal Font

While any unicode capable font can be used with micron, it's highly recommended to use a `*"Nerd Font"`* (see https://www.nerdfonts.com/), which will add a lot of extra glyphs and icons to your output.

> A Few Demo Outputs

`F222`Bddd

`cWith micron, you can control layout and presentation
`a

``

`B33f

You can change background ...

``

`B393

`r`F320... and foreground colors`f
`a

`b

If you want to make a break, horizontal dividers can be inserted. They can be plain, like the one below this text, or you can style them with unicode characters and glyphs, like the wavy divider in the beginning of this document.

-

`cText can be `_underlined`_, `!bold`! or `*italic`*.

You can also `_`*`!`B5d5`F222combine`f`b`_ `_`Ff00f`Ff80o`Ffd0r`F9f0m`F0f2a`F0fdt`F07ft`F43fi`F70fn`Fe0fg`` for some fabulous effects.
`a


>>>Sections and Headings

You can define an arbitrary number of sections and sub sections, each with their own named headings. Text inside sections will be automatically indented.

-

If you place a divider inside a section, it will adhere to the section indents.

>>>>>
If no heading text is defined, the section will appear as a sub-section without a header. This can be useful for creating indented blocks of text, like this one.

>Micron tags

Tags are used to format text with micron. Some tags can appear anywhere in text, and some must appear at the beginning of a line. If you need to write text that contains a sequence that would be interpreted as a tag, you can escape it with the character \.

In the following sections, the different tags will be introduced. Any styling set within micron can be reset to the default style by using the special \`\` tag anywhere in the markup, which will immediately remove any formatting previously specified.

>>Alignment

To control text alignment use the tag \`c to center text, \`l to left-align, \`r to right-align, and \`a to return to the default alignment of the document. Alignment tags must appear at the beginning of a line. Here is an example:

`Faaa
`=
`cThis line will be centered.
So will this.
`aThe alignment has now been returned to default.
`rThis will be aligned to the right
``
`=
``

The above markup produces the following output:

`Faaa`B333

`cThis line will be centered.
So will this.

`aThe alignment has now been returned to default.

`rThis will be aligned to the right

``


>>Formatting

Text can be formatted as `!bold`! by using the \`! tag, `_underline`_ by using the \`_ tag and `*italic`* by using the \`* tag.

Here's an example of formatting text:

`Faaa
`=
We shall soon see `!bold`! paragraphs of text decorated with `_underlines`_ and `*italics`*. Some even dare `!`*`_combine`` them!
`=
``

The above markup produces the following output:

`Faaa`B333

We shall soon see `!bold`! paragraphs of text decorated with `_underlines`_ and `*italics`*. Some even dare `!`*`_combine`!`*`_ them!

``


>>Sections

To create sections and subsections, use the > tag. This tag must be placed at the beginning of a line. To specify a sub-section of any level, use any number of > tags. If text is placed after a > tag, it will be used as a heading.

A line that contains only <code>></code> characters, with no title text (extra spaces after the <code>></code> run still count as "no title"), only updates the depth used to indent what follows. It does not create an empty heading row in the preview.

Here is an example of sections:

`Faaa
`=
>High Level Stuff
This is a section. It contains this text.

>>Another Level
This is a sub section.

>>>Going deeper
A sub sub section. We could continue, but you get the point.

>>>>
Wait! It's worth noting that we can also create sections without headings. They look like this.
`=
``

The above markup produces the following output:

`Faaa`B333
>High Level Stuff
This is a section. It contains this text.

>>Another Level
This is a sub section.

>>>Going deeper
A sub sub section. We could continue, but you get the point.

>>>>
Wait! It's worth noting that we can also create sections without headings. They look like this.
``


For longer pages, a bare subsection line (only \`\>>\` and optional spaces) indents the following block without an extra empty heading in the preview—useful for multi-line layouts (links, toolbars) under a title.

`Faaa
`=
>Portal

>>
`[Node`:/page/index.mu] / more content on the same depth

<
`=
``

>Colors

Foreground colors can be specified with the \`F tag, followed by three hexadecimal characters. To return to the default foreground color, use the \`f tag. Background color is specified in the same way, but by using the \`B and \`b tags.

Here's a few examples:

`Faaa
`=
You can use `B5d5`F222 color `f`b `Ff00f`Ff80o`Ffd0r`F9f0m`F0f2a`F0fdt`F07ft`F43fi`F70fn`Fe0fg`f for some fabulous effects.
`=
``

The above markup produces the following output:

`Faaa`B333

You can use `B5d5`F222 color `f`B333 `Ff00f`Ff80o`Ffd0r`F9f0m`F0f2a`F0fdt`F07ft`F43fi`F70fn`Fe0fg`f for some fabulous effects.

``


>Links

Links to pages, files or other resources can be created with the \`[ tag, which should always be terminated with a closing ]. You can create links with and without labels, it is up to you to control the formatting of links with other tags. Although not strictly necessary, it is good practice to at least format links with underlining.

Here's a few examples:

`Faaa
`=
Here is a link without any label: `[72914442a3689add83a09a767963f57c:/page/index.mu]

This is a `[labeled link`72914442a3689add83a09a767963f57c:/page/index.mu] to the same page, but it's hard to see if you don't know it

Here is `F00a`_`[a more visible link`72914442a3689add83a09a767963f57c:/page/index.mu]`_`f
`=
``

The above markup produces the following output:

`Faaa`B333

Here is a link without any label: `[72914442a3689add83a09a767963f57c:/page/index.mu]

This is a `[labeled link`72914442a3689add83a09a767963f57c:/page/index.mu] to the same page, but it's hard to see if you don't know it

Here is `F00f`_`[a more visible link`72914442a3689add83a09a767963f57c:/page/index.mu]`_`f

``

When links like these are displayed in the built-in browser, clicking on them or activating them using the keyboard will cause the browser to load the specified URL.

>Tables

You can include rendered tables by enclosing them in \`t tags. Optionally, you can also specify alignment and max rendering width by adding these properties to the opening \`t tag, like \`tc30. Here's an example:

`Faaa
`=
`t
| Name | Price | Qty |
| ---- | :---: | --: |
| `F3a3Apple`f | Free | `!5`! |
| Orange | Ask, nicely | 3 |
`t
`=
``

The above markup produces the following table:

`t
| Name | Price | Qty |
| ---- | :---: | --: |
| `F3a3Apple`f | Free | `!5`! |
| Orange | Ask, nicely | 3 |
`t


>Fields & Requests

Nomad Network let's you use simple input fields for submitting data to node-side applications. Submitted data, along with other session variables will be available to the node-side script / program as environment variables.

>>Request Links

Links can contain request variables and a list of fields to submit to the node-side application. You can include all fields on the page, only specific ones, and any number of request variables. To simply submit all fields on a page to a specified node-side page, create a link like this:

`Faaa
`=
`[Submit Fields`:/page/fields.mu`*]
`=
``

Note the `!*`! following the extra `!\``! at the end of the path. This `!*`! denotes `*all fields`*. You can also specify a list of fields to include:

`Faaa
`=
`[Submit Fields`:/page/fields.mu`username|auth_token]
`=
``

If you want to include pre-set variables, you can do it like this:

`Faaa
`=
`[Query the System`:/page/fields.mu`username|auth_token|action=view|amount=64]
`=
``

>> Fields

Here's an example of creating a field. We'll create a field named `!user_input`! and fill it with the text `!Pre-defined data`!. Note that we are using background color tags to make the field more visible to the user:

`Faaa
`=
A simple input field: `B444`<user_input`Pre-defined data>`b
`=
``

You must always set a field `*name`*, but you can of course omit the pre-defined value of the field:

`Faaa
`=
An empty input field: `B444`<demo_empty`>`b
`=
``

You can set the size of the field like this:

`Faaa
`=
A sized input field:  `B444`<16|with_size`>`b
`=
``

It is possible to mask fields, for example for use with passwords and similar:

`Faaa
`=
A masked input field: `B444`<!|masked_demo`hidden text>`b
`=
``

And you can of course control all parameters at the same time:

`Faaa
`=
Full control: `B444`<!32|all_options`hidden text>`b
`=
``

Collecting the above markup produces the following output:

`Faaa`B333

A simple input field: `B444`<user_input`Pre-defined data>`B333

An empty input field: `B444`<demo_empty`>`B333

A sized input field:  `B444`<16|with_size`>`B333

A masked input field: `B444`<!|masked_demo`hidden text>`B333

Full control: `B444`<!32|all_options`hidden text>`B333
`b
>>> Checkboxes

In addition to text fields, Checkboxes are another way of submitting data. They allow the user to make a single selection or select multiple options.

`Faaa
`=
`<?|field_name|value`>`b Label Text`
`=
When the checkbox is checked, it's field will be set to the provided value. If there are multiple checkboxes that share the same field name, the checked values will be concatenated when they are sent to the node by a comma.
``

`B444`<?|sign_up|1`>`b Sign me up`

You can also pre-check both checkboxes and radio groups by appending a |* after the field value.

`B444`<?|checkbox|1|*`>`b Pre-checked checkbox`

>>> Radio groups

Radio groups are another input that lets the user chose from a set of options. Unlike checkboxes, radio buttons with the same field name are mutually exclusive.

Example:

`=
`B900`<^|color|Red`>`b  Red

`B090`<^|color|Green`>`b Green

`B009`<^|color|Blue`>`b Blue
`=

will render:

`B900`<^|color|Red`>`b  Red

`B090`<^|color|Green`>`b Green

`B009`<^|color|Blue`>`b Blue

In this example, when the data is submitted, `B444` field_color`b will be set to whichever value from the list was selected.

``

>Headings with fields

If a line begins with one or more > section markers but also contains a field tag (the sequence \`<...\`>), the heading markers are stripped so the field renders without an extra heading row above it.

`Faaa
`=
>>`B444`<24|demo_user`Ada Lovelace>`b
`=
``

Produces (leading > markers omitted because the line contains a field):

`Faaa`B333

`B444`<24|demo_user`Ada Lovelace>`b


>>True-color (FT / BT)

Six-digit RGB uses \`FT or \`BT followed by six hex digits. The renderer also accepts an interleaved packed form where six nibbles are split around marker bytes (output by some terminal sources). Either form sets the corresponding foreground or background color.

`Faaa
`=
`FTaabbccTruecolor foreground`f `BT334455truecolor background`b
Packed example: `F123`F456hello world``
`=
``

Produces:

`Faaa`B333

`FTaabbccTruecolor foreground`f `BT334455truecolor background`b
Packed example: `F123`F456hello world``


>Partials (embedded fragments)

A line that starts with \`{ declares an embedded fragment. Close it with }. Inside, separated by \`, you can give a destination URL, an optional refresh interval in seconds (must be at least 1), and an optional |-separated list of extra fields. Use pid=NAME inside the fields for a stable identity across refreshes. The renderer emits a placeholder div with class Mu-partial and data-partial-* attributes; the host script (or NomadNet client) is responsible for fetching the destination and replacing the placeholder.

The playground in this page only fetches HTTP-style URLs, so the source below is shown as-is rather than rendered, but the markup demonstrates the three accepted shapes:

`Faaa
`=
`{72914442a3689add83a09a767963f57c:/page/sidebar.mu}
`{72914442a3689add83a09a767963f57c:/page/news.mu`30}
`{72914442a3689add83a09a767963f57c:/page/panel.mu`120`pid=sidebar|theme=dark}
`=
``


>>Line-leading backslash

When the first character of a line is a backslash, it is dropped and the rest of the line is parsed with the next character treated as already-escaped (NomadNet / micron-parser-js parity). Use this when column one must be a literal backtick, > or # rather than a heading or comment.

`Faaa
`=
\> This line does not start a heading.
\# This is not a comment line.
`=
``

Produces:

`Faaa`B333

\> This line does not start a heading.
\# This is not a comment line.


>Comments

You can insert comments that will not be displayed in the output by starting a line with the # character.

Here's an example:

`Faaa
`=
# This line will not be displayed
This line will
`=
``

The above markup produces the following output:

`Faaa`B333

# This line will not be displayed
This line will

``


>Literals

To display literal content, for example source-code, or blocks of text that should not be interpreted by micron, you can use literal blocks, specified by the \`= tag. ASCII art and FIGlet-style banners contain many characters that micron treats specially (for example backticks, underscores, and vertical bars). Put those blocks inside a literal section so each character is shown exactly and monospace alignment is preserved in the preview.

A toggle line may include leading or trailing spaces (the whole line is still recognized as a toggle if it is only \`= and whitespace). Toggles do not add blank rows; only the lines between the toggles are rendered as literal text.

Below is the source code of this entire document, presented as a literal block.

-

`=
        
