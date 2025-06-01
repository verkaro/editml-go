This is a test document for multiline EditML features.

%% A line comment, should be skipped.

First, some multiline inline additions:
This is {+a simple
addition spanning
multiple lines.+ws} an example.

Another one without an ID:
{+This is another
multiline addition.
It even has a blank line

in the middle.+}.

Multiline deletions:
This text is {-gone
and also
spans a few lines.-jd} and should be removed.

And one more:
{-This deletion
is simpler.-}.

Multiline comments:
Here is some text {>this is a
comment that goes
over several lines<xy} that we want to keep.

A comment with an escaped closing operator:
{>This comment contains a literal < character.
It also spans lines.<}.

Multiline highlights:
Let's {=emphasize this
important section
of text.=anon} for review.

Another highlight:
{=This part
is also
highlighted.=}.

Now for structural edits with multiline content.

{move~This is a block of text
that is intended to be moved.
It has multiple lines.
It even contains an escaped tilde: ~ here.
~moveML1}

Some text in between.

{copy~This is a block
to be copied.
Line 1
Line 2
Line 3~copyML1}

More text.

And here are the targets:
The moved text should appear here: {move:moveML1}.

The copied text should appear here: {copy:copyML1}.
And also here: {copy:copyML1}.

Mixing it up:
This is {+an added line
that also contains {-a nested
deletion (which should be literal text per spec 3.3.4)-}
and more added text.+dev}.

A structural move with inline edits inside its content:
{move~This block will be moved.
It contains {+added text within the block+ws}
and also some {-deleted text within the block-ws}.
This should all move together.
~moveML2}

The target for the mixed content move:
{move:moveML2}

End of multiline tests.
