---
title: Epson scanner, large PDFs and ImageMagick (convert CLI)
date: 2017-09-12
---

The Epson iPrint soft is producing PDFs in 300dpi but the 300 information does
not seem to be stored in the PDF metadata; thus, the PDF reader thinks
by default that it is a 72 dpi PDF and thus the huge size:

    87,49 × 123,72 cm

instead of an A4 paper:

    21 × 29,7 cm

To convert this 300 dpi into a real 300 dpi:

    gs -sDEVICE=pdfwrite -dCompatibilityLevel=1.4 -dPDFSETTINGS=/ebook -dNOPAUSE -dQUIET -dBATCH -dDetectDuplicateImages -dCompressFonts=true -sPAPERSIZE=a4 -dPDFFitPage -sOutputFile=output.pdf input.pdf

- /screen selects low-resolution output similar to the Acrobat Distiller (up to version X) "Screen Optimized" setting.
- /ebook selects medium-resolution output similar to the Acrobat Distiller (up to version X) "eBook" setting.
- /printer selects output similar to the Acrobat Distiller "Print Optimized" (up to version X) setting.
- /prepress selects output similar to Acrobat Distiller "Prepress Optimized" (up to version X) setting.
- /default selects output intended to be useful across a wide variety of uses, possibly at the expense of a larger output file.

See:

    open /usr/local/share/ghostscript/9.22/doc/index.html

---

    (WORKS BUT BIGGER SIZE AFTERWARDS)
    convert -units PixelsPerInch -density 300 filein.pdf fileout.pdf
    convert -units PixelsPerInch EPSON004.pdf -density 300 -quality 30 -compress jpeg fileout.pdf

---

## Color to grayscale

From: <https://superuser.com/questions/104656/convert-a-pdf-to-greyscale-on-the-command-line-in-floss>

```shell
while read i; do
    gs -sDEVICE=pdfwrite -sColorConversionStrategy=Gray -dProcessColorModel=/DeviceGray -dCompatibilityLevel=1.4 -dNOPAUSE -dBATCH -sOutputFile="${i/.pdf/_gray.pdf}" "$i";
done <<< "some.pdf"
```

Using ImageMagick (mogrify):

```shell
mogrify -despeckle -fuzz 5% -fill white -opaque white -gamma 0.8 -colorspace gray -depth 6 -format png \*.jpg
```

## OCR

## To black and white

A SO link talking about that:
<https://stackoverflow.com/questions/15211428/conversion-of-tiff-to-pdf-with-ghostscript>.

```shell
gs -dQUIET -dNOPAUSE -r200 -dBATCH -sPAPERSIZE=a4 -sDEVICE=tiffg3 -sOutputFile=temp.tiff -sColorConversionStrategy=Mono -sColorConversionStrategyForImages=/Mono -dProcessColorModel=/DeviceGray "a.pdf"
tiff2pdf -o out.pdf -p A4 -F temp.tiff
```

Or:

```shell
pdfimages -j "some.pdf" out/
cd out/
mogrify -despeckle -fuzz 5% -fill white -opaque white -gamma 0.8 -colorspace gray -depth 6 -format png \*.jpg
```

## Remove noise created by scanning text in JPEG

```shell
pdfimages -j "some.pdf" out/
cd out/
mogrify -despeckle -fuzz 5% -fill white -opaque white -gamma 0.8 -colorspace gray -depth 6 -format png *.jpg
```
