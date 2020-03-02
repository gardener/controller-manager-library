<p>Packages:</p>
<ul>
<li>
<a href="#example.examples.gardener.cloud%2fv1alpha1">example.examples.gardener.cloud/v1alpha1</a>
</li>
</ul>
<h2 id="example.examples.gardener.cloud/v1alpha1">example.examples.gardener.cloud/v1alpha1</h2>
<p>
<p>Package v1alpha1 contains example API resources.</p>
</p>
Resource Types:
<ul><li>
<a href="#example.examples.gardener.cloud/v1alpha1.Example">Example</a>
</li></ul>
<h3 id="example.examples.gardener.cloud/v1alpha1.Example">Example
</h3>
<p>
<p>Example is an example for a custom resource.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>apiVersion</code></br>
string</td>
<td>
<code>
example.examples.gardener.cloud/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code></br>
string
</td>
<td><code>Example</code></td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#example.examples.gardener.cloud/v1alpha1.ExampleSpec">
ExampleSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>hostname</code></br>
<em>
string
</em>
</td>
<td>
<p>Hostname is a host name</p>
</td>
</tr>
<tr>
<td>
<code>scheme</code></br>
<em>
string
</em>
</td>
<td>
<p>URLScheme is an URL scheme name to compose an url</p>
</td>
</tr>
<tr>
<td>
<code>port</code></br>
<em>
int
</em>
</td>
<td>
<p>Port is a port name for the URL</p>
</td>
</tr>
<tr>
<td>
<code>path</code></br>
<em>
string
</em>
</td>
<td>
<p>Path is a path for the URL</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="example.examples.gardener.cloud/v1alpha1.ExampleSpec">ExampleSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#example.examples.gardener.cloud/v1alpha1.Example">Example</a>)
</p>
<p>
<p>ExampleSpec is  the specification for an example object.</p>
</p>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>hostname</code></br>
<em>
string
</em>
</td>
<td>
<p>Hostname is a host name</p>
</td>
</tr>
<tr>
<td>
<code>scheme</code></br>
<em>
string
</em>
</td>
<td>
<p>URLScheme is an URL scheme name to compose an url</p>
</td>
</tr>
<tr>
<td>
<code>port</code></br>
<em>
int
</em>
</td>
<td>
<p>Port is a port name for the URL</p>
</td>
</tr>
<tr>
<td>
<code>path</code></br>
<em>
string
</em>
</td>
<td>
<p>Path is a path for the URL</p>
</td>
</tr>
</tbody>
</table>
<hr/>
