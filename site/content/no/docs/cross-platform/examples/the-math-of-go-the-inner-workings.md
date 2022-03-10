
---
title: "Matten til Go: Drivverket"
linkTitle: "Matten til Go"
date: 2017-01-05
description: >
 En kort oppsummering av denne siden. Tekst kan **utheves** sller skrives i _kursiv_ og kan ha flere avsnitt.
---

Text can be **bold**, _italic_, or ~~strikethrough~~. [Links](https://gohugo.io) should be blue with no underlines (unless hovered over).

There should be whitespace between paragraphs. Lorem markdownum tempus auras formasque ore vir crescere est! Malo quod, hunc, est dura; aut haec simillima nec per conantemque iusserat audax moriensque confessasque. Haec vulneret quam libratum homo pede arbore tu manus membrisque iuveni Clymeneia se cepi unda, iustae? Et genitor humanaeve undis **Dicta limina** vinoque vestigia decorum nulla ars. Pectora sede: quoque magnum Persidaque in suos, adiciunt tenebor.

Formidine humo velle vulnera remotis admonitu suo mora vivo ubi. Libidine et mittor Orphei nulla. Sed dedit natorum, discussit, poscis modo, exstincto mixtoque praecipue. Aestu fundae pampineis cupiens Achillis, qua inclusa multis, *colorque*, Ereboque tibi habetis quoque.


> There should be no margin above this first sentence.
>
> Blockquotes should be a lighter gray with a border along the left side in the secondary color.
>
> There should be no margin below this final sentence.

## First Header 2

This is a normal paragraph following a header. Victrix tridentigero corripuere messibus, una rector, me se Iovis. *Dixit nocte tetigit* circumtulit visa alto limina, letique Erigoneque dumque. Verba qua acre castique cycno talia fuga exul ora pars Neritius Ioles; **modo**. Solacia fores servat querno tamen! Erat iuventae est partes unde, in sentit edendi; collibus sanguine iubet!

Deae legum paulatimque terra, non vos mutata tacet: dic. Vocant docuique me plumas fila quin afuerunt copia haec o neque.

On big screens, paragraphs and headings should not take up the full container width, but we want tables, code blocks and similar to take the full width.

Erat vera cur scelus mundo quam? Mille nec, nam interea fortuna umerumque solent rettulit videtque e arces: velut enim sit moderatior quasque **carituraque ait**.


## Second Header 2

> This is a blockquote following a header. Bacon ipsum dolor sit amet t-bone doner shank drumstick, pork belly porchetta chuck sausage brisket ham hock rump pig. Chuck kielbasa leberkas, pork bresaola ham hock filet mignon cow shoulder short ribs biltong.

### Header 3

```
This is a code block following a header.
```

Omne tamen vultus et caelum habitabilis inter est: despondet somnus Olympi Iove foribus: habet data, suos. Suis illi auro verba sibi os Turno. Oris avis mariti callida deficis tangor.


#### Header 4

* This is an unordered list following a header.
* This is an unordered list following a header.
* This is an unordered list following a header.

##### Header 5

1. This is an ordered list following a header.
2. This is an ordered list following a header.
3. This is an ordered list following a header.

###### Header 6

| What      | Follows         |
|-----------|-----------------|
| A table   | A header        |
| A table   | A header        |
| A table   | A header        |

----------------

There's a horizontal rule above and below this.

----------------

Here is an unordered list:

* Liverpool F.C.
* Chelsea F.C.
* Manchester United F.C.

And an ordered list:

1. Michael Brecker
2. Seamus Blake
3. Branford Marsalis

And an unordered task list:

- [x] Create a Hugo theme
- [x] Add task lists to it
- [ ] Take a vacation

And a "mixed" task list:

- [ ] Pack bags
- ?
- [ ] Travel!

And a nested list:

* Jackson 5
  * Michael
  * Tito
  * Jackie
  * Marlon
  * Jermaine
* TMNT
  * Leonardo
  * Michelangelo
  * Donatello
  * Raphael

Definition lists can be used with Markdown syntax. Definition headers are bold.

Name
: Godzilla

Born
: 1952

Birthplace
: Japan

Color
: Green


----------------

Tables should have bold headings and alternating shaded rows.

| Artist            | Album           | Year |
|-------------------|-----------------|------|
| Michael Jackson   | Thriller        | 1982 |
| Prince            | Purple Rain     | 1984 |
| Beastie Boys      | License to Ill  | 1986 |

If a table is too wide, it should scroll horizontally.

| Artist            | Album           | Year | Label       | Awards   | Songs     |
|-------------------|-----------------|------|-------------|----------|-----------|
| Michael Jackson   | Thriller        | 1982 | Epic Records | Grammy Award for Album of the Year, American Music Award for Favorite Pop/Rock Album, American Music Award for Favorite Soul/R&B Album, Brit Award for Best Selling Album, Grammy Award for Best Engineered Album, Non-Classical | Wanna Be Startin' Somethin', Baby Be Mine, The Girl Is Mine, Thriller, Beat It, Billie Jean, Human Nature, P.Y.T. (Pretty Young Thing), The Lady in My Life |
| Prince            | Purple Rain     | 1984 | Warner Brothers Records | Grammy Award for Best Score Soundtrack for Visual Media, American Music Award for Favorite Pop/Rock Album, American Music Award for Favorite Soul/R&B Album, Brit Award for Best Soundtrack/Cast Recording, Grammy Award for Best Rock Performance by a Duo or Group with Vocal | Let's Go Crazy, Take Me With U, The Beautiful Ones, Computer Blue, Darling Nikki, When Doves Cry, I Would Die 4 U, Baby I'm a Star, Purple Rain |
| Beastie Boys      | License to Ill  | 1986 | Mercury Records | noawardsbutthistablecelliswide | Rhymin & Stealin, The New Style, She's Crafty, Posse in Effect, Slow Ride, Girls, (You Gotta) Fight for Your Right, No Sleep Till Brooklyn, Paul Revere, Hold It Now, Hit It, Brass Monkey, Slow and Low, Time to Get Ill |

----------------

Code snippets like `var foo = "bar";` can be shown inline.

Also, `this should vertically align` ~~`with this`~~ ~~and this~~.

Code can also be shown in a block element.

```
foo := "bar";
bar := "foo";
```

Code can also use syntax highlighting.

```go
func main() {
  input := `var foo = "bar";`

  lexer := lexers.Get("javascript")
  iterator, _ := lexer.Tokenise(nil, input)
  style := styles.Get("github")
  formatter := html.New(html.WithLineNumbers())

  var buff bytes.Buffer
  formatter.Format(&buff, style, iterator)

  fmt.Println(buff.String())
}
```

```
Long, single-line code blocks should not wrap. They should horizontally scroll if they are too long. This line should be long enough to demonstrate this.
```

Inline code inside table cells should still be distinguishable.

| Language    | Code               |
|-------------|--------------------|
| Javascript  | `var foo = "bar";` |
| Ruby        | `foo = "bar"{`      |

----------------

Small images should be shown at their actual size.

![](https://upload.wikimedia.org/wikipedia/commons/thumb/9/9e/Picea_abies_shoot_with_buds%2C_Sogndal%2C_Norway.jpg/240px-Picea_abies_shoot_with_buds%2C_Sogndal%2C_Norway.jpg)

Large images should always scale down and fit in the content container.

![](https://upload.wikimedia.org/wikipedia/commons/thumb/9/9e/Picea_abies_shoot_with_buds%2C_Sogndal%2C_Norway.jpg/1024px-Picea_abies_shoot_with_buds%2C_Sogndal%2C_Norway.jpg)

_The photo above of the Spruce Picea abies shoot with foliage buds: Bjørn Erik Pedersen, CC-BY-SA._


## Components

### Alerts

{{< alert >}}This is an alert.{{< /alert >}}
{{< alert title="Note" >}}This is an alert with a title.{{< /alert >}}
{{% alert title="Note" %}}This is an alert with a title and **Markdown**.{{% /alert %}}
{{< alert color="success" >}}This is a successful alert.{{< /alert >}}
{{< alert color="warning" >}}This is a warning.{{< /alert >}}
{{< alert color="warning" title="Warning" >}}This is a warning with a title.{{< /alert >}}


## Another Heading

Add some sections here to see how the ToC looks like. Bacon ipsum dolor sit amet t-bone doner shank drumstick, pork belly porchetta chuck sausage brisket ham hock rump pig. Chuck kielbasa leberkas, pork bresaola ham hock filet mignon cow shoulder short ribs biltong.

### This Document

Inguina genus: Anaphen post: lingua violente voce suae meus aetate diversi. Orbis unam nec flammaeque status deam Silenum erat et a ferrea. Excitus rigidum ait: vestro et Herculis convicia: nitidae deseruit coniuge Proteaque adiciam *eripitur*? Sitim noceat signa *probat quidem*. Sua longis *fugatis* quidem genae.


### Pixel Count

Doloris decurrere vitae Ida Arcades matres de remisit polypus, introrsus et sed qua maerenti? Serpit meta illic ut sinu. Transformat ungues genitor, et visis ademit sustinet abstulit lampadibus. Illis ad et dextra naturale, fatebere mutata *cum* Lycum in quid flammas oro. Populus Aurora caerula et feremus clavigeri ungues dubitant et inde corpore clamat, qui non Ilioneus pugnat abstuleris undas, habet.

### Contact Info

Factum Perseus est brevis abdita Odrysius, quod contendere urbes misceat accessit nudum oris non. Cumque dentibus nullam nec mille potentia regnumque supplex!


### External Links

Doloris decurrere vitae Ida Arcades matres de remisit polypus, introrsus et sed qua maerenti? Serpit meta illic ut sinu. Transformat ungues genitor, et visis ademit sustinet abstulit lampadibus. Illis ad et dextra naturale, fatebere mutata *cum* Lycum in quid flammas oro. Populus Aurora caerula et feremus clavigeri ungues dubitant et inde corpore clamat, qui non Ilioneus pugnat abstuleris undas, habet.



```
This is the final element on the page and there should be no margin below this.
```